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

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRuntimeEnvFindsByteman(t *testing.T) {
	if os.Getenv("CHAOSD_ENV_HELPER") == "1" {
		if err := SetRuntimeEnv(); err != nil {
			t.Fatal(err)
		}
		got, err := exec.Command("bminstall.sh").Output()
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "byteman-found" {
			t.Fatalf("unexpected script output %q", got)
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range []bool{false, true} {
		name := "bundled"
		if preset {
			name = "preset"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			helper := filepath.Join(dir, "env.test")
			if err := os.WriteFile(helper, binary, 0700); err != nil {
				t.Fatal(err)
			}
			home := filepath.Join(dir, "tools", "byteman")
			if preset {
				home = filepath.Join(dir, "custom-byteman")
			}
			if err := os.MkdirAll(filepath.Join(dir, "tools"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(home, "bin"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, "bin", "bminstall.sh"), []byte("#!/bin/sh\nprintf byteman-found"), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("BYTEMAN_HOME", "")
			if preset {
				t.Setenv("BYTEMAN_HOME", home)
			}
			t.Setenv("CHAOSD_ENV_HELPER", "1")
			if output, err := exec.Command(helper, "-test.run=^TestRuntimeEnvFindsByteman$").CombinedOutput(); err != nil {
				t.Fatalf("helper: %v\n%s", err, output)
			}
		})
	}
}
