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

package config

import "testing"

func TestContainerRuntimeValidation(t *testing.T) {
	for _, runtime := range []string{"docker", "containerd"} {
		conf := Config{Platform: LocalPlatform, Runtime: runtime}
		if err := conf.Validate(); err != nil {
			t.Errorf("runtime %s: %v", runtime, err)
		}
	}
	conf := Config{Platform: LocalPlatform, Runtime: "runtime"}
	if err := conf.Validate(); err == nil {
		t.Fatal("accepted unknown runtime")
	}
}
