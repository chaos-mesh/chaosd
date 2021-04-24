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

	"github.com/go-logr/zapr"
	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	"go.uber.org/zap"

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
	var newRun *core.ExperimentRun
	defer func() {
		var updErr error
		if panicErr, ok := recover().(error); panicErr != nil && ok {
			log.Error("scheduled run errored", zap.String("expId", cj.experiment.Uid), zap.Error(panicErr))
			if newRun != nil {
				updErr = cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunFailed, panicErr.Error())
			}
		} else {
			log.Info("scheduled run success", zap.String("expId", cj.experiment.Uid))
			if newRun != nil {
				updErr = cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunSuccess, "")
			} else {
				log.Info("cron run finished")
			}
		}
		if updErr != nil {
			log.Error("failed to update experiment run", zap.Error(updErr))
		}
	}()

	log.Info("Started new run", zap.String("expId", cj.experiment.Uid))
	cfg, err := cj.experiment.GetRequestCommand()
	if err != nil {
		panic(perr.WithStack(err))
	}
	cronDuration, err := cfg.ScheduleDuration()
	if err != nil {
		panic(perr.WithStack(err))
	}
	if cronDuration != nil && time.Until(cj.experiment.CreatedAt.Add(*cronDuration)).Milliseconds() < 0 {
		if err = cj.scheduler.Remove(cj.experiment.ID); err != nil {
			panic(perr.WithStack(err))
		}
		if err = cj.scheduler.expStore.Update(context.Background(), cj.experiment.Uid, core.Success, "", cj.experiment.RecoverCommand); err != nil {
			panic(perr.WithStack(err))
		}
		return
	}

	newRun = cj.experiment.NewRun()
	if err = cj.scheduler.expRunStore.NewRun(context.Background(), newRun); err != nil {
		panic(perr.WithStack(err))
	}

	log.Info("executing attack on new exp run", zap.String("expRunUID", newRun.UID))
	if err = cj.run(); err != nil {
		panic(perr.WithMessage(err, "attack failed"))
	}
}

func NewScheduler(expRunStore core.ExperimentRunStore, expStore core.ExperimentStore) Scheduler {
	return Scheduler{
		Cron: cron.New(
			cron.WithLocation(time.UTC),
			cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
			cron.WithLogger(zapr.NewLogger(log.L())),
		),
		expRunStore: expRunStore,
		expStore:    expStore,
		cronStore:   &cronStore{entry: make(map[uint]cron.EntryID)},
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
	defer func() {
		log.Info("Scheduled new attack cron", zap.String("expUid", exp.Uid), zap.String("cron", spec))
	}()
	return nil
}

func (scheduler Scheduler) Remove(expId uint) error {
	scheduler.Cron.Remove(scheduler.cronStore.Remove(expId))
	return nil
}

func (scheduler Scheduler) Start() {
	log.Info("starting Scheduler")
	scheduler.Cron.Start()
}
