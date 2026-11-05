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
