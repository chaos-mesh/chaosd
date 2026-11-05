// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/scheduler"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
	"github.com/chaos-mesh/chaosd/pkg/store/dbstore"
	"github.com/chaos-mesh/chaosd/pkg/store/experiment"
)

func newAttackRouter(t *testing.T) (*gin.Engine, core.ExperimentStore) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "experiments.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	store := experiment.NewStore(&dbstore.DB{DB: db})
	chaos := chaosd.NewServer(nil, store, nil, nil, nil, nil, nil, scheduler.Scheduler{})
	router := gin.New()
	router.Use(utils.MWHandleErrors())
	NewServer(nil, chaos, store).handler(router)
	return router, store
}

func requestAttack(t *testing.T, router *gin.Engine, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("status %d, want %d: %s", response.Code, status, response.Body.String())
	}
	return response
}

type fileRouteAttack struct {
	delivered core.FileCommand
	recovered core.FileCommand
}

func (f *fileRouteAttack) Attack(options core.AttackConfig, _ chaosd.Environment) error {
	command := options.(*core.FileCommand)
	f.delivered = *command
	if command.Action == core.FileModifyPrivilegeAction {
		command.OriginPrivilege = 640
	}
	return nil
}

func (f *fileRouteAttack) Recover(exp core.Experiment, _ chaosd.Environment) error {
	return json.Unmarshal([]byte(exp.RecoverCommand), &f.recovered)
}

func TestFileAttackRoute(t *testing.T) {
	original := chaosd.FileAttack
	attack := &fileRouteAttack{}
	chaosd.FileAttack = attack
	t.Cleanup(func() { chaosd.FileAttack = original })
	router, _ := newAttackRouter(t)

	for _, body := range []string{"{", `{"action":"append","file-name":"file"}`} {
		requestAttack(t, router, http.MethodPost, "/api/attack/file", body, http.StatusBadRequest)
	}
	if attack.delivered.Action != "" {
		t.Fatalf("invalid request reached file attack: %+v", attack.delivered)
	}

	response := requestAttack(t, router, http.MethodPost, "/api/attack/file",
		`{"uid":"file-append","kind":"process","action":"append","file-name":"appended","data":"tail"}`, http.StatusOK)
	var result utils.Response
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.UID != "file-append" || attack.delivered.Kind != core.FileAttack || attack.delivered.Count != 1 {
		t.Fatalf("UID = %q, delivered = %#v", result.UID, attack.delivered)
	}

	requestAttack(t, router, http.MethodPost, "/api/attack/file",
		`{"uid":"file-modify","action":"modify","file-name":"modified","privilege":600}`, http.StatusOK)
	requestAttack(t, router, http.MethodDelete, "/api/attack/file-modify", "", http.StatusOK)
	recovered := attack.recovered
	if recovered.UID != "file-modify" || recovered.Kind != core.FileAttack || recovered.Action != core.FileModifyPrivilegeAction ||
		recovered.FileName != "modified" || recovered.Privilege != 600 || recovered.OriginPrivilege != 640 {
		t.Fatalf("recovery received wrong file command: %#v", recovered)
	}
}

type hostRouteManager struct {
	actions []string
	err     error
}

func (*hostRouteManager) Name() string { return "test" }

func (h *hostRouteManager) Shutdown() error {
	h.actions = append(h.actions, core.HostShutdownAction)
	return h.err
}

func (h *hostRouteManager) Reboot() error {
	h.actions = append(h.actions, core.HostRebootAction)
	return h.err
}

func TestHostAttackRoute(t *testing.T) {
	original := chaosd.Host
	host := &hostRouteManager{}
	chaosd.Host = host
	t.Cleanup(func() { chaosd.Host = original })
	router, store := newAttackRouter(t)

	for _, body := range []string{"{", `{"action":"unsupported"}`} {
		requestAttack(t, router, http.MethodPost, "/api/attack/host", body, http.StatusBadRequest)
	}
	if len(host.actions) != 0 {
		t.Fatalf("invalid request invoked host actions: %v", host.actions)
	}

	for _, action := range []string{core.HostShutdownAction, core.HostRebootAction} {
		t.Run(action, func(t *testing.T) {
			host.actions = nil
			uid := "host-" + action
			body := fmt.Sprintf(`{"action":%q,"uid":%q,"kind":"process"}`, action, uid)
			response := requestAttack(t, router, http.MethodPost, "/api/attack/host", body, http.StatusOK)
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != uid || len(host.actions) != 1 || host.actions[0] != action {
				t.Fatalf("UID = %q, host actions = %v, want %q", result.UID, host.actions, action)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Kind != core.HostAttack {
				t.Fatalf("persisted kind = %q, want host", exp.Kind)
			}
		})
	}

	response := requestAttack(t, router, http.MethodDelete, "/api/attack/host-reboot", "", http.StatusInternalServerError)
	if !strings.Contains(response.Body.String(), "non_recoverable_attack") {
		t.Fatalf("missing non-recoverable error: %s", response.Body.String())
	}
	exp, err := store.FindByUid(context.Background(), "host-reboot")
	if err != nil {
		t.Fatal(err)
	}
	if exp.Status != core.Success || len(host.actions) != 1 {
		t.Fatalf("failed recovery changed status %q or invoked host actions %v", exp.Status, host.actions)
	}

	host.err = errors.New("host backend failed")
	requestAttack(t, router, http.MethodPost, "/api/attack/host",
		`{"uid":"host-error","action":"shutdown"}`, http.StatusInternalServerError)
}
