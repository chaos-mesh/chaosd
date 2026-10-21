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
