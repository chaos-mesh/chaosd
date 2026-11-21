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

package core

import "testing"

func TestKafkaCommandString(t *testing.T) {
	for _, action := range []KafkaAttackAction{KafkaFillAction, KafkaFloodAction, KafkaIOAction} {
		command := NewKafkaCommand()
		command.Action = action
		if got := command.String(); got != string(action) {
			t.Errorf("action %q: String() = %q", action, got)
		}
	}
}

func TestKafkaScheduleDurationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, schedule, duration string
		invalid                  bool
	}{
		{"unscheduled", "", "", false},
		{"scheduled with duration", "@every 1h", "10s", false},
		{"scheduled invalid duration", "@every 1h", "not-a-duration", true},
		{"scheduled empty duration", "@every 1h", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := NewKafkaCommand()
			command.Action = KafkaFloodAction
			command.Topic = "orders"
			command.Host = "broker"
			command.Port = 9092
			command.MessageSize = 128
			command.Threads = 1
			command.Schedule = tc.schedule
			command.Duration = tc.duration
			if err := command.Validate(); (err != nil) != tc.invalid {
				t.Errorf("schedule=%q duration=%q: %v", tc.schedule, tc.duration, err)
			}
		})
	}
}
