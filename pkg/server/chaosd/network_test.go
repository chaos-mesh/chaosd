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
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type failingTCQueryStore struct {
	core.TCRuleStore
	err error
}

func (*failingTCQueryStore) DeleteByExperiment(context.Context, string) error {
	return nil
}

func (s *failingTCQueryStore) FindByDevice(context.Context, string) ([]*core.TCRule, error) {
	return nil, s.err
}

func TestRecoverTCQueryFailure(t *testing.T) {
	queryErr := errors.New("TC rule query failed")
	// A nil daemon makes any SetTcs call fail before it can change host rules.
	server := &Server{tcRule: &failingTCQueryStore{err: queryErr}}
	if err := server.recoverTC(uuid.NewString(), "eth0"); !errors.Is(err, queryErr) {
		t.Fatalf("recoverTC error = %v, want query error", err)
	}
}
