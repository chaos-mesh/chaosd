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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

func TestHTTPProxyHelper(t *testing.T) {
	capture := os.Getenv("CHAOSD_PROXY_CAPTURE")
	if capture == "" {
		return
	}
	request, err := http.ReadRequest(bufio.NewReader(os.Stdin))
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != http.MethodPut {
		t.Fatalf("method = %s", request.Method)
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(capture, body, 0600); err != nil {
		t.Fatal(err)
	}
	fmt.Fprint(os.Stdout, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
}

func TestHTTPProxyRoute(t *testing.T) {
	for _, action := range []string{"delay", "abort"} {
		t.Run(action, func(t *testing.T) {
			dir := t.TempDir()
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("CHAOSD_HTTP_HELPER", binary)
			t.Setenv("CHAOSD_PROXY_CAPTURE", filepath.Join(dir, "request.json"))
			t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
			if err := os.WriteFile(filepath.Join(dir, "tproxy"), []byte("#!/bin/sh\nexec \"$CHAOSD_HTTP_HELPER\" -test.run=^TestHTTPProxyHelper$\n"), 0700); err != nil {
				t.Fatal(err)
			}
			db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "proxy.db")), &gorm.Config{})
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
			body := `{"action":"abort","target":"Request","proxy_ports":[8080],"uid":"proxy-test"}`
			if action == "delay" {
				body = `{"action":"delay","target":"Request","proxy_ports":[8080],"delay":"250ms","uid":"proxy-test"}`
			}
			request := httptest.NewRequest(http.MethodPost, "/api/attack/http", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != "proxy-test" {
				t.Fatalf("UID = %q", result.UID)
			}
			exp, err := store.FindByUid(context.Background(), result.UID)
			if err != nil {
				t.Fatal(err)
			}
			var saved core.HTTPAttackConfig
			if err := json.Unmarshal([]byte(exp.RecoverCommand), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.ProxyPID <= 0 {
				t.Fatal("missing proxy PID")
			}
			t.Cleanup(func() {
				var status syscall.WaitStatus
				if _, err := syscall.Wait4(saved.ProxyPID, &status, 0, nil); err != nil {
					t.Error(err)
				} else if !status.Exited() || status.ExitStatus() != 0 {
					t.Errorf("proxy helper status: %v", status)
				}
			})
			if exp.Kind != core.HTTPAttack || exp.Status != core.Success || exp.LaunchMode != core.ServerMode {
				t.Fatalf("unexpected persisted experiment: %+v", exp)
			}
			capture, err := os.ReadFile(filepath.Join(dir, "request.json"))
			if err != nil {
				t.Fatal(err)
			}
			var delivered core.HTTPAttackConfig
			if err := json.Unmarshal(capture, &delivered.Config); err != nil {
				t.Fatal(err)
			}
			if len(delivered.Config.ProxyPorts) != 1 || delivered.Config.ProxyPorts[0] != 8080 {
				t.Fatalf("ports = %v", delivered.Config.ProxyPorts)
			}
			if len(delivered.Config.Rules) != 1 {
				t.Fatalf("rules = %+v", delivered.Config.Rules)
			}
			rule := delivered.Config.Rules[0]
			if action == "delay" && (rule.Actions.Delay == nil || *rule.Actions.Delay != "250ms") {
				t.Fatalf("delay was not delivered: %+v", rule.Actions)
			}
			if action == "abort" && (rule.Actions.Abort == nil || !*rule.Actions.Abort) {
				t.Fatalf("abort was not delivered: %+v", rule.Actions)
			}
			if string(delivered.Config.Rules[0].Target) != "Request" {
				t.Fatalf("target = %s", delivered.Config.Rules[0].Target)
			}
		})
	}
}
