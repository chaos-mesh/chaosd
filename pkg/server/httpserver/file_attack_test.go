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
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
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
	f.recovered = core.FileCommand{}
	return json.Unmarshal([]byte(exp.RecoverCommand), &f.recovered)
}

func TestFileAttackRoute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "file.db")), &gorm.Config{})
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
	original := chaosd.FileAttack
	attack := &fileRouteAttack{}
	chaosd.FileAttack = attack
	t.Cleanup(func() { chaosd.FileAttack = original })
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
		{"malformed", "{"},
		{"unsupported", `{"action":"unsupported"}`},
		{"create target", `{"action":"create"}`},
		{"modify file", `{"action":"modify","privilege":600}`},
		{"modify privilege", `{"action":"modify","file-name":"file"}`},
		{"delete target", `{"action":"delete"}`},
		{"rename source", `{"action":"rename","dest-file":"new"}`},
		{"rename destination", `{"action":"rename","source-file":"old"}`},
		{"append file", `{"action":"append","data":"tail"}`},
		{"append data", `{"action":"append","file-name":"file"}`},
		{"replace file", `{"action":"replace","origin-string":"old","dest-string":"new"}`},
		{"replace origin", `{"action":"replace","file-name":"file","dest-string":"new"}`},
		{"replace destination", `{"action":"replace","file-name":"file","origin-string":"old"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request(t, http.MethodPost, "/api/attack/file", tc.body, http.StatusBadRequest)
			exps, err := store.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(exps) != 0 {
				t.Fatalf("invalid request persisted %d experiments", len(exps))
			}
		})
	}
	for _, tc := range []struct {
		action, fields string
		want           core.FileCommand
	}{
		{"create", `"file-name":"created","kind":"process"`, core.FileCommand{FileName: "created"}},
		{"modify", `"file-name":"modified","privilege":600`, core.FileCommand{FileName: "modified", Privilege: 600}},
		{"delete", `"dir-name":"deleted"`, core.FileCommand{DirName: "deleted"}},
		{"rename", `"source-file":"old","dest-file":"new"`, core.FileCommand{SourceFile: "old", DestFile: "new"}},
		{"append", `"file-name":"appended","data":"tail"`, core.FileCommand{FileName: "appended", Data: "tail", Count: 1}},
		{"replace", `"file-name":"replaced","origin-string":"old","dest-string":"new","line":2`, core.FileCommand{FileName: "replaced", OriginStr: "old", DestStr: "new", Line: 2}},
	} {
		t.Run(tc.action, func(t *testing.T) {
			uid := "file-" + tc.action
			body := fmt.Sprintf(`{"action":%q,"uid":%q,%s}`, tc.action, uid, tc.fields)
			response := request(t, http.MethodPost, "/api/attack/file", body, http.StatusOK)
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != uid {
				t.Fatalf("UID = %q, want %q", result.UID, uid)
			}
			tc.want.CommonAttackConfig = core.CommonAttackConfig{Action: tc.action, Kind: core.FileAttack, UID: uid}
			if !reflect.DeepEqual(attack.delivered, tc.want) {
				t.Fatalf("delivered %#v, want %#v", attack.delivered, tc.want)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Kind != core.FileAttack || exp.Status != core.Success || exp.LaunchMode != core.ServerMode {
				t.Fatalf("unexpected persisted experiment: %+v", exp)
			}
			if tc.action == core.FileModifyPrivilegeAction {
				tc.want.OriginPrivilege = 640
			}
			var saved core.FileCommand
			if err := json.Unmarshal([]byte(exp.RecoverCommand), &saved); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(saved, tc.want) {
				t.Fatalf("saved %#v, want %#v", saved, tc.want)
			}
			request(t, http.MethodDelete, "/api/attack/"+uid, "", http.StatusOK)
			if !reflect.DeepEqual(attack.recovered, tc.want) {
				t.Fatalf("recovered %#v, want %#v", attack.recovered, tc.want)
			}
			exp, err = store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Status != core.Destroyed {
				t.Fatalf("recovered status = %q", exp.Status)
			}
		})
	}
}
