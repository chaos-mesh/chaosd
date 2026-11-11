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

type multiRecoveryStore struct {
	core.ExperimentStore
	exps  []*core.Experiment
	conds *core.SearchCommand
}

func (s *multiRecoveryStore) ListByConditions(_ context.Context, conds *core.SearchCommand) ([]*core.Experiment, error) {
	s.conds = conds
	return s.exps, nil
}

//revive:disable-next-line:var-naming Match the ExperimentStore interface.
func (s *multiRecoveryStore) FindByUid(_ context.Context, uid string) (*core.Experiment, error) {
	for _, exp := range s.exps {
		if exp.Uid == uid {
			return exp, nil
		}
	}
	return nil, nil
}

func (s *multiRecoveryStore) Update(_ context.Context, uid, status, _, _ string) error {
	exp, _ := s.FindByUid(context.Background(), uid)
	exp.Status = status
	return nil
}

func TestRecoverAttacksContinuesAfterFailures(t *testing.T) {
	portOccupied := core.NetworkCommand{
		CommonAttackConfig: core.CommonAttackConfig{Kind: core.NetworkAttack, Action: core.NetworkPortOccupiedAction},
		PortPid:            int32(exitedPID(t)),
	}
	store := &multiRecoveryStore{exps: []*core.Experiment{
		{Uid: "unknown-kind", Kind: "unknown", Status: core.Success},
		{Uid: "host", Kind: core.HostAttack, Status: core.Success},
		{Uid: "network", Kind: core.NetworkAttack, Status: core.Success, RecoverCommand: portOccupied.RecoverData()},
	}}
	server := &Server{expStore: store}

	results, err := server.RecoverAttacks("")
	if err != nil {
		t.Fatal(err)
	}
	if store.conds.Kind != "" || store.conds.Status != core.Success || store.conds.Asc {
		t.Fatalf("listed experiments with %+v, want newest experiments of any kind in the success status", store.conds)
	}
	if len(results) != len(store.exps) {
		t.Fatalf("got %d results, want %d", len(results), len(store.exps))
	}

	if results[0].Err == nil || errorx.IsOfType(results[0].Err, core.ErrNonRecoverableAttack) {
		t.Errorf("unknown kind: got error %v, want a recovery failure", results[0].Err)
	}
	if !errorx.IsOfType(results[1].Err, core.ErrNonRecoverableAttack) {
		t.Errorf("host: got error %v, want a non-recoverable error", results[1].Err)
	}
	if results[2].Err != nil {
		t.Errorf("network: %v", results[2].Err)
	}

	for i, want := range []string{core.Success, core.Success, core.Destroyed} {
		if exp := store.exps[i]; exp.Status != want {
			t.Errorf("%s status %q, want %q", exp.Uid, exp.Status, want)
		}
	}
}
