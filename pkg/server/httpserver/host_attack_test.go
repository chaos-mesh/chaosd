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

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "host.db")), &gorm.Config{})
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
	server := NewServer(nil, chaos, store)
	router := gin.New()
	router.Use(utils.MWHandleErrors())
	server.handler(router)
	request := func(t *testing.T, method, path, body string, status int) *httptest.ResponseRecorder {
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
	for _, tc := range []struct{ name, body string }{
		{"empty body", ""},
		{"malformed", "{"},
		{"missing action", "{}"},
		{"empty action", `{"action":""}`},
		{"unsupported", `{"action":"unsupported"}`},
		{"invalid schedule duration", `{"action":"reboot","schedule":"* * * * *","duration":"invalid"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request(t, http.MethodPost, "/api/attack/host", tc.body, http.StatusBadRequest)
			if len(host.actions) != 0 {
				t.Fatalf("invalid request invoked host actions: %v", host.actions)
			}
			exps, err := store.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(exps) != 0 {
				t.Fatalf("invalid request persisted %d experiments", len(exps))
			}
		})
	}
	for _, action := range []string{core.HostShutdownAction, core.HostRebootAction} {
		t.Run(action, func(t *testing.T) {
			host.actions = nil
			uid := "host-" + action
			body := fmt.Sprintf(`{"action":%q,"uid":%q,"kind":"process"}`, action, uid)
			response := request(t, http.MethodPost, "/api/attack/host", body, http.StatusOK)
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != uid {
				t.Fatalf("UID = %q, want %q", result.UID, uid)
			}
			if len(host.actions) != 1 || host.actions[0] != action {
				t.Fatalf("host actions = %v, want [%s]", host.actions, action)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Uid != uid || exp.Kind != core.HostAttack || exp.Action != action || exp.Status != core.Success || exp.LaunchMode != core.ServerMode {
				t.Fatalf("unexpected persisted experiment: %+v", exp)
			}
			var saved core.HostCommand
			if err := json.Unmarshal([]byte(exp.RecoverCommand), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.UID != uid || saved.Kind != core.HostAttack || saved.Action != action {
				t.Fatalf("unexpected saved host command: %+v", saved)
			}
			response = request(t, http.MethodDelete, "/api/attack/"+uid, "", http.StatusInternalServerError)
			if !strings.Contains(response.Body.String(), "non_recoverable_attack") {
				t.Fatalf("missing non-recoverable error: %s", response.Body.String())
			}
			exp, err = store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Status != core.Success {
				t.Fatalf("failed recovery changed status to %q", exp.Status)
			}
			if len(host.actions) != 1 {
				t.Fatalf("recovery invoked host actions: %v", host.actions)
			}
		})
	}
	for _, action := range []string{core.HostShutdownAction, core.HostRebootAction} {
		t.Run(action+" backend error", func(t *testing.T) {
			host.actions = nil
			host.err = errors.New("host backend failed")
			uid := "host-error-" + action
			body := fmt.Sprintf(`{"action":%q,"uid":%q}`, action, uid)
			request(t, http.MethodPost, "/api/attack/host", body, http.StatusInternalServerError)
			if len(host.actions) != 1 || host.actions[0] != action {
				t.Fatalf("host actions = %v, want [%s]", host.actions, action)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Kind != core.HostAttack || exp.Status != core.Error || exp.LaunchMode != core.ServerMode {
				t.Fatalf("unexpected failed experiment: %+v", exp)
			}
		})
	}
}
