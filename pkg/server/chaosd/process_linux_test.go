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

package chaosd

import (
	"os"
	"strconv"
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func TestProcessAttackResetsPIDs(t *testing.T) {
	pid := os.Getpid()
	options := &core.ProcessCommand{
		Process: strconv.Itoa(pid),
		Signal:  0, // Signal 0 checks existence without delivering a signal.
	}

	for run := 1; run <= 2; run++ {
		if err := ProcessAttack.Attack(options, Environment{}); err != nil {
			t.Fatalf("run %d: attack failed: %v", run, err)
		}
		if len(options.PIDs) != 1 || options.PIDs[0] != pid {
			t.Fatalf("run %d: PIDs = %v, want [%d]", run, options.PIDs, pid)
		}
	}
}
