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

func (s *recoveryStore) Update(_ context.Context, _, status, _, _ string) error {
	s.updated = true
	s.exp.Status = status
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

func TestRecoverPortOccupiedWithExitedHelperProcess(t *testing.T) {
	config := core.NetworkCommand{
		CommonAttackConfig: core.CommonAttackConfig{Kind: core.NetworkAttack, Action: core.NetworkPortOccupiedAction},
		PortPid:            int32(exitedPID(t)),
	}
	store := &recoveryStore{exp: &core.Experiment{
		Uid:            "exited-helper",
		Kind:           config.AttackKind(),
		Status:         core.Success,
		RecoverCommand: config.RecoverData(),
	}}
	server := &Server{expStore: store}
	if err := server.RecoverAttack("exited-helper"); err != nil {
		t.Fatalf("recover with exited helper process: %v", err)
	}
	if store.exp.Status != core.Destroyed {
		t.Fatalf("experiment status %q, want %q", store.exp.Status, core.Destroyed)
	}
}
