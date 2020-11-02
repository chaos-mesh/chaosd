// Copyright 2020 Chaos Mesh Authors.
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

package experiment

import (
	"context"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
	"github.com/chaos-mesh/chaos-daemon/pkg/store/dbstore"
)

func NewStore(db *dbstore.DB) core.ExperimentStore {
	db.AutoMigrate(&core.Experiment{})

	es := &experimentStore{db}

	return es
}

type experimentStore struct {
	db *dbstore.DB
}

func (e *experimentStore) List(_ context.Context) ([]*core.Experiment, error) {
	return nil, nil
}

func (e *experimentStore) ListByStatus(_ context.Context, status string) ([]*core.Experiment, error) {
	return nil, nil
}

func (e *experimentStore) FindByUid(_ context.Context, uid string) (*core.Experiment, error) {
	return nil, nil
}

func (e *experimentStore) Set(_ context.Context, exp *core.Experiment) error {
	return nil
}

func (e *experimentStore) Update(_ context.Context, uid, status, msg string, pids []int) error {
	return nil
}
