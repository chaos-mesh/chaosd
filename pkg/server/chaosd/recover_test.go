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
	"context"
	"os/exec"
	"testing"

	"github.com/joomcode/errorx"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type recoveryStore struct {
	core.ExperimentStore
	exp     *core.Experiment
	updated bool
}

//revive:disable-next-line:var-naming Match the ExperimentStore interface.
func (s *recoveryStore) FindByUid(context.Context, string) (*core.Experiment, error) {
	return s.exp, nil
}

func (s *recoveryStore) Update(context.Context, string, string, string, string) error {
	s.updated = true
	return nil
}

func TestNonRecoverableAttackReturnsError(t *testing.T) {
	store := &recoveryStore{exp: &core.Experiment{Uid: "host-test", Kind: core.HostAttack, Status: core.Success}}
	server := &Server{expStore: store}
	err := server.RecoverAttack("host-test")
	if !errorx.IsOfType(err, core.ErrNonRecoverableAttack) {
		t.Fatalf("expected non-recoverable error, got %v", err)
	}
	if store.updated {
		t.Fatal("non-recoverable attack changed recorded status")
	}
}

// exitedPID returns the PID of a child process that has already exited and been reaped.
func exitedPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("run helper process: %v", err)
	}
	return cmd.Process.Pid
}

func TestRecoverAttackWithExitedHelperProcess(t *testing.T) {
	pid := exitedPID(t)
	httpConfig := core.HTTPAttackConfig{
		CommonAttackConfig: core.CommonAttackConfig{Kind: core.HTTPAttack, Action: core.HTTPAbortAction},
		ProxyPID:           pid,
	}
	portConfig := core.NetworkCommand{
		CommonAttackConfig: core.CommonAttackConfig{Kind: core.NetworkAttack, Action: core.NetworkPortOccupiedAction},
		PortPid:            int32(pid),
	}

	tests := []struct {
		name   string
		kind   string
		config core.AttackConfig
	}{
		{name: "http", kind: core.HTTPAttack, config: &httpConfig},
		{name: "port occupied", kind: core.NetworkAttack, config: &portConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &recoveryStore{exp: &core.Experiment{
				Uid:            "exited-helper",
				Kind:           tt.kind,
				Status:         core.Success,
				RecoverCommand: tt.config.RecoverData(),
			}}
			server := &Server{expStore: store}
			if err := server.RecoverAttack("exited-helper"); err != nil {
				t.Fatalf("recover with exited helper process: %v", err)
			}
			if !store.updated {
				t.Fatal("recovery did not mark the experiment as destroyed")
			}
		})
	}
}
