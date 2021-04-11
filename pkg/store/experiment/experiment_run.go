// Copyright 2021 Chaos Mesh Authors.
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
	"errors"

	perr "github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/store/dbstore"
)

func NewRunStore(db *dbstore.DB) core.ExperimentRunStore {
	db.AutoMigrate(&core.ExperimentRun{})
	es := &experimentRunStore{db}
	return es
}

type experimentRunStore struct {
	db *dbstore.DB
}

func (store *experimentRunStore) ListByExperimentID(ctx context.Context, id uint) ([]*core.ExperimentRun, error) {
	runs := make([]*core.ExperimentRun, 0)
	if err := store.db.
		Preload("Experiment").
		Find(&runs, "experiment_id = ?", id).
		Order("start_at DESC").
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perr.WithStack(err)
	}

	return runs, nil
}

func (store *experimentRunStore) ListByExperimentUID(ctx context.Context, uid string) ([]*core.ExperimentRun, error) {
	runs := make([]*core.ExperimentRun, 0)
	if err := store.db.
		Joins("JOIN experiments ON experiments.uid = ?", uid).
		Find(&runs).
		Order("start_at DESC").
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perr.WithStack(err)
	}

	return runs, nil
}

func (store *experimentRunStore) LatestRun(ctx context.Context, id uint) (*core.ExperimentRun, error) {
	run := &core.ExperimentRun{}
	if err := store.db.
		Preload("Experiment").
		Order("start_at DESC").
		First(&run, "experiment_id = ?", id).
		Error; err != nil {
		return nil, perr.WithStack(err)
	}

	return run, nil
}

func (store *experimentRunStore) NewRun(_ context.Context, expRun *core.ExperimentRun) error {
	return store.db.Model(core.ExperimentRun{}).Save(expRun).Error
}

func (store *experimentRunStore) Update(_ context.Context, runUid string, status string, message string) error {
	return store.db.
		Model(core.ExperimentRun{}).
		Where("uid = ?", runUid).
		Updates(core.Experiment{Status: status, Message: message}).
		Error
}
