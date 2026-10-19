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
	"strings"
	"sync/atomic"
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

func TestHTTPAttackRoute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "http.db")), &gorm.Config{})
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
	var requests int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"malformed", "{", http.StatusBadRequest},
		{"invalid", `{"action":"delay","target":"Request","proxy_ports":[80],"delay":"invalid"}`, http.StatusBadRequest},
		{"request", fmt.Sprintf(`{"action":"request","url":%q,"count":1,"uid":"http-test"}`, target.URL), http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/attack/http", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status == http.StatusOK {
				var body utils.Response
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.UID != "http-test" {
					t.Fatalf("UID = %q", body.UID)
				}
			}
		})
	}
	if atomic.LoadInt32(&requests) != 1 {
		t.Errorf("target received %d requests, want 1", requests)
	}
	exps, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(exps) != 1 {
		t.Fatalf("persisted %d experiments, want 1", len(exps))
	}
	if exps[0].Kind != core.HTTPAttack || exps[0].Status != core.Success || exps[0].LaunchMode != core.ServerMode {
		t.Fatalf("unexpected persisted experiment: %+v", exps[0])
	}
}
