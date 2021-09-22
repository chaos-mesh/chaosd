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

package core

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	RunStarted   = "started"
	RunFailed    = "failed"
	RunSuccess   = "success"
	RunRecovered = "recovered"
)

// ExperimentRunStore defines operations for working with experiment runs
type ExperimentRunStore interface {
	ListByExperimentID(ctx context.Context, id uint) ([]*ExperimentRun, error)
	ListByExperimentUID(ctx context.Context, uid string) ([]*ExperimentRun, error)
	LatestRun(ctx context.Context, id uint) (*ExperimentRun, error)

	NewRun(ctx context.Context, expRun *ExperimentRun) error
	Update(ctx context.Context, runUid string, status string, message string) error
}

// ExperimentRun represents a run of an experiment
type ExperimentRun struct {
	ID           uint      `gorm:"primary_key" json:"id"`
	UID          string    `gorm:"index:uid" json:"uid"`
	StartAt      time.Time `gorm:"autoCreateTime" json:"start_at"`
	FinishedAt   time.Time `json:"finished_at"`
	Status       string    `json:"status"`
	Message      string    `json:"error"`
	ExperimentID uint
	Experiment   Experiment `gorm:"foreignKey:ExperimentID" json:"experiment"`
}

func (exp Experiment) NewRun() *ExperimentRun {
	return &ExperimentRun{
		ExperimentID: exp.ID,
		// TODO: maybe need to use specified uid
		UID:    uuid.New().String(),
		Status: RunStarted,
	}
}
