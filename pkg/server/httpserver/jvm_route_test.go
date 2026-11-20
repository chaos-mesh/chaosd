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
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

type jvmRouteAttack struct {
	calls     int
	delivered core.JVMCommand
}

func (j *jvmRouteAttack) Attack(options core.AttackConfig, _ chaosd.Environment) error {
	j.calls++
	j.delivered = *options.(*core.JVMCommand)
	return nil
}

func (j *jvmRouteAttack) Recover(core.Experiment, chaosd.Environment) error { return nil }

func TestJVMAttackRoutePersistsAction(t *testing.T) {
	original := chaosd.JVMAttack
	attack := &jvmRouteAttack{}
	chaosd.JVMAttack = attack
	t.Cleanup(func() { chaosd.JVMAttack = original })
	router, store := newAttackRouter(t)

	for _, tt := range []struct {
		action string
		fields string
	}{
		{core.JVMGCAction, ""},
		{core.JVMLatencyAction, `,"class":"Order","method":"pay","latency":100`},
		{core.JVMStressAction, `,"cpu-count":1`},
	} {
		t.Run(tt.action, func(t *testing.T) {
			uid := "jvm-" + tt.action
			body := fmt.Sprintf(`{"uid":%q,"action":%q,"pid":1234%s}`, uid, tt.action, tt.fields)
			before := attack.calls
			response := requestAttack(t, router, http.MethodPost, "/api/attack/jvm", body, http.StatusOK)
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != uid || attack.calls != before+1 || attack.delivered.Action != tt.action {
				t.Fatalf("UID = %q, calls = %d, delivered = %#v", result.UID, attack.calls-before, attack.delivered)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Action != tt.action {
				t.Fatalf("persisted action = %q, want %q", exp.Action, tt.action)
			}
		})
	}
}
