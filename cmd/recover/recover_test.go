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

package recover

import "testing"

func TestRecoverCommandValidate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options recoverCommand
		invalid bool
	}{
		{"uid", recoverCommand{uid: "uid"}, false},
		{"all", recoverCommand{all: true}, false},
		{"kind", recoverCommand{kind: "network"}, false},
		{"nothing", recoverCommand{}, true},
		{"uid with all", recoverCommand{uid: "uid", all: true}, true},
		{"uid with kind", recoverCommand{uid: "uid", kind: "network"}, true},
		{"unknown kind", recoverCommand{kind: "unknown"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.options.validate(); (err != nil) != tc.invalid {
				t.Fatalf("invalid=%v, got error %v", tc.invalid, err)
			}
		})
	}
}
