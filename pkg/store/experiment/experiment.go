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
	"errors"

	"gorm.io/gorm"

	perr "github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/store/dbstore"
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
	exps := make([]*core.Experiment, 0)
	if err := e.db.
		Find(&exps).
		Order("created_at DESC").
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perr.WithStack(err)
	}

	return exps, nil
}

func (e *experimentStore) ListByStatus(_ context.Context, status string) ([]*core.Experiment, error) {
	exps := make([]*core.Experiment, 0)
	if err := e.db.
		Where("status = ?", status).
		Find(&exps).
		Order("created_at DESC").
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perr.WithStack(err)
	}

	return exps, nil
}

func (e *experimentStore) ListByConditions(_ context.Context, conds *core.SearchCommand) ([]*core.Experiment, error) {
	if conds == nil {
		return nil, errors.New("conditions is required")
	}

	exps := make([]*core.Experiment, 0)

	db := e.db.Model(core.Experiment{})

	if conds.Offset > 0 {
		db = db.Offset(int(conds.Offset))
	}

	if conds.Limit > 0 {
		db = db.Limit(int(conds.Limit))
	}

	if !conds.All {
		if len(conds.Type) > 0 {
			db = db.Where("kind = ?", conds.Type)
		}

		if len(conds.Status) > 0 {
			db = db.Where("status = ?", conds.Status)
		}
	}

	order := "create_at"
	if !conds.Asc {
		order += " DESC"
	}

	if err := db.
		Find(&exps).
		Order(order).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, perr.WithStack(err)
	}

	return exps, nil
}

func (e *experimentStore) FindByUid(_ context.Context, uid string) (*core.Experiment, error) {
	exps := make([]*core.Experiment, 0)
	if err := e.db.
		Where("uid = ?", uid).
		Find(&exps).
		Order("created_at DESC").
		Error; err != nil {
		return nil, perr.WithStack(err)
	}

	if len(exps) > 0 {
		return exps[0], nil
	}

	return nil, gorm.ErrRecordNotFound
}

func (e *experimentStore) Set(_ context.Context, exp *core.Experiment) error {
	return e.db.Model(core.Experiment{}).Save(exp).Error
}

func (e *experimentStore) Update(_ context.Context, uid, status, msg string, command string) error {
	return e.db.
		Model(core.Experiment{}).
		Where("uid = ?", uid).
		Updates(core.Experiment{Status: status, Message: msg, RecoverCommand: command}).
		Error
}
