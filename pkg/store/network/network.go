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

package network

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
	"github.com/chaos-mesh/chaos-daemon/pkg/store/dbstore"
)

func NewIPSetRuleStore(db *dbstore.DB) core.IPSetRuleStore {
	db.AutoMigrate(&core.IPSetRule{})

	is := &ipsetRuleStore{db}

	return is
}

type ipsetRuleStore struct {
	db *dbstore.DB
}

func (i *ipsetRuleStore) List(_ context.Context) ([]*core.IPSetRule, error) {
	rules := make([]*core.IPSetRule, 0)
	if err := i.db.
		Find(&rules).
		Order("created_at DESC").
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}

	return rules, nil
}

func (i *ipsetRuleStore) Set(_ context.Context, rule *core.IPSetRule) error {
	return i.db.Model(core.IPSetRule{}).Save(rule).Error
}

func (i *ipsetRuleStore) FindByExperiment(_ context.Context, experiment string) ([]*core.IPSetRule, error) {
	rules := make([]*core.IPSetRule, 0)
	if err := i.db.
		Where("experiment = ?", experiment).
		Find(&rules).Order("created_at DESC").
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}

	return rules, nil
}

func (i *ipsetRuleStore) DeleteByExperiment(_ context.Context, experiment string) error {
	return i.db.
		Where("experiment = ?", experiment).
		Unscoped().
		Delete(core.IPSetRule{}).
		Error
}

func NewIptablesRuleStore(db *dbstore.DB) core.IptablesRuleStore {
	db.AutoMigrate(&core.IptablesRule{})

	is := &iptablesRuleStore{db}

	return is
}

type iptablesRuleStore struct {
	db *dbstore.DB
}

func (i *iptablesRuleStore) List(_ context.Context) ([]*core.IptablesRule, error) {
	rules := make([]*core.IptablesRule, 0)
	if err := i.db.
		Find(&rules).
		Order("created_at DESC").
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}
	return rules, nil
}

func (i *iptablesRuleStore) Set(_ context.Context, rule *core.IptablesRule) error {
	return i.db.Model(core.IptablesRule{}).Save(rule).Error
}

func (i *iptablesRuleStore) FindByExperiment(_ context.Context, experiment string) ([]*core.IptablesRule, error) {
	rules := make([]*core.IptablesRule, 0)
	if err := i.db.
		Where("experiment = ?", experiment).
		Find(&rules).
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}

	return rules, nil
}

func (i *iptablesRuleStore) DeleteByExperiment(_ context.Context, experiment string) error {
	return i.db.
		Where("experiment = ?", experiment).
		Unscoped().
		Delete(core.IptablesRule{}).
		Error
}

func NewTCRuleStore(db *dbstore.DB) core.TCRuleStore {
	db.AutoMigrate(&core.TCRule{})

	ts := &tcRuleStore{db}

	return ts
}

type tcRuleStore struct {
	db *dbstore.DB
}

func (t *tcRuleStore) Set(_ context.Context, rule *core.TCRule) error {
	return t.db.Model(core.TCRule{}).Save(rule).Error
}

func (t *tcRuleStore) List(_ context.Context) ([]*core.TCRule, error) {
	rules := make([]*core.TCRule, 0)
	if err := t.db.
		Find(&rules).
		Order("created_at DESC").
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}
	return rules, nil
}

func (t *tcRuleStore) FindByExperiment(_ context.Context, experiment string) ([]*core.TCRule, error) {
	rules := make([]*core.TCRule, 0)
	if err := t.db.
		Where("experiment = ?", experiment).
		Find(&rules).
		Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, errors.WithStack(err)
	}
	return rules, nil
}

func (t *tcRuleStore) DeleteByExperiment(_ context.Context, experiment string) error {
	return t.db.
		Where("experiment = ?", experiment).
		Unscoped().
		Delete(core.TCRule{}).
		Error
}
