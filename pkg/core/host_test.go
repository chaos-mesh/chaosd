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

func TestHostCommandValidate(t *testing.T) {
	for _, tc := range []struct {
		name, action, schedule, duration string
		invalid                          bool
	}{
		{name: "shutdown", action: HostShutdownAction},
		{name: "reboot", action: HostRebootAction},
		{name: "empty action", invalid: true},
		{name: "unsupported action", action: "unsupported", invalid: true},
		{name: "scheduled", action: HostRebootAction, schedule: "* * * * *", duration: "1m"},
		{name: "missing duration", action: HostShutdownAction, schedule: "* * * * *", invalid: true},
		{name: "invalid duration", action: HostRebootAction, schedule: "* * * * *", duration: "invalid", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := NewHostCommand()
			command.Action = tc.action
			command.Schedule = tc.schedule
			command.Duration = tc.duration
			if err := command.Validate(); (err != nil) != tc.invalid {
				t.Fatalf("Validate() = %v, want invalid=%t", err, tc.invalid)
			}
		})
	}
}
