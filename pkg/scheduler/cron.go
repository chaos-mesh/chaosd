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

package scheduler

import (
	"context"
	"time"

	perr "github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type Scheduler struct {
	*cron.Cron
	expStore    core.ExperimentStore
	expRunStore core.ExperimentRunStore
	cronStore   CronStore
}

type CronJob struct {
	scheduler  *Scheduler
	experiment *core.Experiment
	run        func() error
}

// TODO: write tests for it
func (cj *CronJob) Run() {
	var err error
	cfg, err := cj.experiment.GetRequestCommand()
	if err != nil {
		panic(perr.WithStack(err))
	}
	cronDuration, err := cfg.ScheduleDuration()
	if err != nil {
		panic(perr.WithStack(err))
	}
	if cj.experiment.CreatedAt.Add(cronDuration).Sub(time.Now()).Microseconds() >= 0 {
		cj.scheduler.Remove(cj.experiment.ID)
		if err = cj.scheduler.expStore.Update(context.Background(), cj.experiment.Uid, core.Success, "", cj.experiment.RecoverCommand); err != nil {
			panic(perr.WithStack(err))
		}
		return
	}

	newRun := cj.experiment.NewRun()
	if err = cj.scheduler.expRunStore.NewRun(context.Background(), newRun); err != nil {
		panic(perr.WithStack(err))
	}

	defer func() {
		if err != nil {
			cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunFailed, err.Error())
		} else {
			cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunSuccess, "")
		}
	}()

	if err = cj.run(); err != nil {
		panic(perr.WithMessage(err, "attack failed"))
	}
}

func NewScheduler(expRunStore core.ExperimentRunStore, expStore core.ExperimentStore) Scheduler {
	return Scheduler{
		Cron: cron.New(
			cron.WithLocation(time.UTC),
			cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
		),
		expRunStore: expRunStore,
		expStore:    expStore,
		cronStore:   &cronStore{entry: make(map[uint]cron.EntryID, 0)},
	}
}

func (scheduler *Scheduler) Schedule(exp *core.Experiment, spec string, task func() error) error {
	cj := CronJob{
		scheduler:  scheduler,
		experiment: exp,
		run:        task,
	}
	entryId, err := scheduler.AddJob(spec, &cj)
	if err != nil {
		return err
	}
	scheduler.cronStore.Add(exp.ID, entryId)
	return nil
}

func (scheduler Scheduler) Remove(expId uint) error {
	scheduler.Cron.Remove(scheduler.cronStore.Remove(expId))
	return nil
}
