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
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

// pidAttack records a different PID on each run, like attacks that start a process.
type pidAttack struct {
	nextPid   int32
	recovered []int32
}

func (a *pidAttack) Attack(options core.AttackConfig, _ Environment) error {
	a.nextPid++
	options.(*core.StressCommand).StressngPid = a.nextPid
	return nil
}

func (a *pidAttack) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	a.recovered = append(a.recovered, config.(*core.StressCommand).StressngPid)
	return nil
}

func TestScheduledRunRecoversAttackState(t *testing.T) {
	options := core.NewStressCommand()
	options.Action = core.StressCPUAction
	exp := &core.Experiment{Uid: "scheduled-test", Kind: options.AttackKind(), RecoverCommand: options.RecoverData()}
	// The scheduler caches the request command decoded from the creation data.
	if _, err := exp.GetRequestCommand(); err != nil {
		t.Fatal(err)
	}

	attack := &pidAttack{}
	attackFunc, recoverFunc := scheduledRunFuncs(attack, options, exp, Environment{AttackUid: exp.Uid})
	for range 2 {
		if err := attackFunc(); err != nil {
			t.Fatal(err)
		}
		if err := recoverFunc(); err != nil {
			t.Fatal(err)
		}
	}

	if len(attack.recovered) != 2 || attack.recovered[0] != 1 || attack.recovered[1] != 2 {
		t.Fatalf("recovered PIDs %v, want [1 2]", attack.recovered)
	}
}
