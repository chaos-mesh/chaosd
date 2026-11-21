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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

// newScheduledAttackRouter builds the attack routes on a real cron scheduler, so a request that
// reaches ExecuteAttack with a cron spec is registered instead of panicking on a nil scheduler.
func newScheduledAttackRouter(t *testing.T) *gin.Engine {
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
	runStore := experiment.NewRunStore(&dbstore.DB{DB: db})
	chaos := chaosd.NewServer(nil, store, runStore, nil, nil, nil, nil, scheduler.NewScheduler(runStore, store))
	router := gin.New()
	router.Use(utils.MWHandleErrors())
	NewServer(nil, chaos, store).handler(router)
	return router
}

type recordingAttack struct{ calls int }

func (r *recordingAttack) Attack(core.AttackConfig, chaosd.Environment) error {
	r.calls++
	return nil
}

func (r *recordingAttack) Recover(core.Experiment, chaosd.Environment) error { return nil }

// TestScheduledAttackRejectsInvalidDuration checks that every kind validates the scheduler
// duration at admission. A scheduled attack whose duration does not parse is accepted and marked
// scheduled, and then every cron tick fails inside CronJob.Run, so the fault is never injected.
func TestScheduledAttackRejectsInvalidDuration(t *testing.T) {
	config := filepath.Join(t.TempDir(), "server.properties")
	if err := os.WriteFile(config, []byte("log.dirs=/var/lib/kafka"), 0600); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		kind, path, fields string
	}{
		// process already validates the scheduler duration and is the control for this table.
		{core.ProcessAttack, "/api/attack/process", `"action":"kill","process":"java"`},
		{core.KafkaAttack, "/api/attack/kafka", fmt.Sprintf(`"action":"flood","topic":"orders","host":"broker","port":9092,"messageSize":128,"threads":1,"configFile":%q`, config)},
		{core.JVMAttack, "/api/attack/jvm", `"action":"gc","pid":1234`},
	} {
		t.Run(tt.kind, func(t *testing.T) {
			original := map[string]chaosd.AttackType{
				core.ProcessAttack: chaosd.ProcessAttack,
				core.KafkaAttack:   chaosd.KafkaAttack,
				core.JVMAttack:     chaosd.JVMAttack,
			}
			attack := &recordingAttack{}
			chaosd.ProcessAttack, chaosd.KafkaAttack, chaosd.JVMAttack = attack, attack, attack
			t.Cleanup(func() {
				chaosd.ProcessAttack = original[core.ProcessAttack]
				chaosd.KafkaAttack = original[core.KafkaAttack]
				chaosd.JVMAttack = original[core.JVMAttack]
			})
			router := newScheduledAttackRouter(t)

			for _, duration := range []string{"not-a-duration", ""} {
				body := fmt.Sprintf(`{"uid":"sched-%s","schedule":"@every 1h","duration":%q,%s}`, tt.kind, duration, tt.fields)
				requestAttack(t, router, http.MethodPost, tt.path, body, http.StatusBadRequest)
			}
			if attack.calls != 0 {
				t.Fatalf("unschedulable request reached the attack %d times", attack.calls)
			}
		})
	}
}
