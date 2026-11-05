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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

type kafkaRouteAttack struct {
	calls     int
	delivered core.KafkaCommand
	recovered core.KafkaCommand
}

func (k *kafkaRouteAttack) Attack(options core.AttackConfig, _ chaosd.Environment) error {
	command := options.(*core.KafkaCommand)
	k.calls++
	k.delivered = *command
	switch command.Action {
	case core.KafkaFillAction:
		command.OriginConfig = "log.retention.bytes=1024\n"
	case core.KafkaIOAction:
		command.OriginModeOfFiles["/var/lib/kafka/orders-0/000.log"] = 0640
	}
	return nil
}

func (k *kafkaRouteAttack) Recover(exp core.Experiment, _ chaosd.Environment) error {
	k.recovered = core.KafkaCommand{}
	return json.Unmarshal([]byte(exp.RecoverCommand), &k.recovered)
}

func TestKafkaAttackRoute(t *testing.T) {
	original := chaosd.KafkaAttack
	attack := &kafkaRouteAttack{}
	chaosd.KafkaAttack = attack
	t.Cleanup(func() { chaosd.KafkaAttack = original })
	router, store := newAttackRouter(t)
	config := filepath.Join(t.TempDir(), "server.properties")
	if err := os.WriteFile(config, []byte("log.dirs=/var/lib/kafka"), 0600); err != nil {
		t.Fatal(err)
	}

	for _, body := range []string{
		`{"topic":"orders","action":"unsupported"}`,
		`{"topic":"orders","action":"flood","host":"broker","port":9092,"messageSize":128}`,
		`{"topic":"orders","action":"fill"}`,
		fmt.Sprintf(`{"topic":"orders","action":"io","configFile":%q}`, config),
		`{"topic":"orders","action":"flood","username":"worker"}`,
	} {
		requestAttack(t, router, http.MethodPost, "/api/attack/kafka", body, http.StatusBadRequest)
	}
	if attack.calls != 0 {
		t.Fatalf("invalid request reached Kafka attack %d times", attack.calls)
	}

	for _, tt := range []struct {
		action string
		fields string
	}{
		{"flood", `"host":"broker","port":9092,"messageSize":128,"threads":1,"username":"worker","password":"secret","authMechanism":"sasl/plain"`},
		{"fill", fmt.Sprintf(`"host":"broker","port":9092,"messageSize":128,"maxBytes":4096,"reloadCommand":"reload-kafka","configFile":%q`, config)},
		{"io", fmt.Sprintf(`"host":"broker","port":9092,"configFile":%q,"nonReadable":true`, config)},
	} {
		t.Run(tt.action, func(t *testing.T) {
			uid := "kafka-" + tt.action
			body := fmt.Sprintf(`{"uid":%q,"topic":"orders","action":%q,%s}`, uid, tt.action, tt.fields)
			before := attack.calls
			response := requestAttack(t, router, http.MethodPost, "/api/attack/kafka", body, http.StatusOK)
			var result utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.UID != uid || attack.calls != before+1 || string(attack.delivered.Action) != tt.action {
				t.Fatalf("UID = %q, calls = %d, delivered = %#v", result.UID, attack.calls-before, attack.delivered)
			}
			exp, err := store.FindByUid(context.Background(), uid)
			if err != nil {
				t.Fatal(err)
			}
			if exp.Action != tt.action || exp.Kind != core.KafkaAttack || exp.Status != core.Success || exp.LaunchMode != core.ServerMode {
				t.Fatalf("wrong persisted experiment: %#v", exp)
			}
			if strings.Count(exp.RecoverCommand, `"action":`) != 1 || strings.Contains(exp.RecoverCommand, `"Action":`) {
				t.Fatalf("recovery data must contain one canonical action: %s", exp.RecoverCommand)
			}
			requestAttack(t, router, http.MethodDelete, "/api/attack/"+uid, "", http.StatusOK)
			recovered := attack.recovered
			if recovered.UID != uid || recovered.Kind != core.KafkaAttack || string(recovered.Action) != tt.action || recovered.Topic != "orders" {
				t.Fatalf("wrong recovery command: %#v", recovered)
			}
			if tt.action == "fill" && (recovered.ConfigFile != config || recovered.OriginConfig != "log.retention.bytes=1024\n") {
				t.Fatalf("lost original Kafka config: %#v", recovered)
			}
			if tt.action == "io" && recovered.OriginModeOfFiles["/var/lib/kafka/orders-0/000.log"] != 0640 {
				t.Fatalf("lost original Kafka file modes: %#v", recovered)
			}
		})
	}
}

func TestKafkaLegacyRecoveryRoute(t *testing.T) {
	original := chaosd.KafkaAttack
	attack := &kafkaRouteAttack{}
	chaosd.KafkaAttack = attack
	t.Cleanup(func() { chaosd.KafkaAttack = original })
	router, store := newAttackRouter(t)

	for _, tt := range []struct {
		action core.KafkaAttackAction
		data   string
	}{
		{core.KafkaFillAction, `{"schedule":"","duration":"","action":"","kind":"kafka","uid":"legacy-fill","Action":"fill","reloadCommand":"reload-kafka","configFile":"/etc/kafka/server.properties","originConfig":"log.retention.bytes=1024\n"}`},
		{core.KafkaIOAction, `{"schedule":"","duration":"","action":"","kind":"kafka","uid":"legacy-io","Action":"io","originModeOfFiles":{"/var/lib/kafka/orders-0/000.log":416}}`},
	} {
		t.Run(string(tt.action), func(t *testing.T) {
			uid := "legacy-" + string(tt.action)
			exp := &core.Experiment{Uid: uid, Kind: core.KafkaAttack, Status: core.Success, RecoverCommand: tt.data}
			if err := store.Set(context.Background(), exp); err != nil {
				t.Fatal(err)
			}
			requestAttack(t, router, http.MethodDelete, "/api/attack/"+uid, "", http.StatusOK)
			recovered := attack.recovered
			if recovered.UID != uid || recovered.Action != tt.action {
				t.Fatalf("wrong legacy recovery command: %#v", recovered)
			}
			if tt.action == core.KafkaFillAction && (recovered.ConfigFile != "/etc/kafka/server.properties" || recovered.ReloadCommand != "reload-kafka" || recovered.OriginConfig != "log.retention.bytes=1024\n") {
				t.Fatalf("lost legacy config: %#v", recovered)
			}
			if tt.action == core.KafkaIOAction && recovered.OriginModeOfFiles["/var/lib/kafka/orders-0/000.log"] != 0640 {
				t.Fatalf("lost legacy file modes: %#v", recovered)
			}
		})
	}
}
