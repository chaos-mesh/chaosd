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
	"sync"
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
	// immutable fields
	scheduler   *Scheduler
	experiment  *core.Experiment
	attackFunc  func() error
	recoverFunc func() error

	// mutable fields protected by sync.Locker
	waitForRecovery bool
	lock            sync.Mutex
}

func hasCronDurationExceeded(startedAt time.Time, duration time.Duration) bool {
	return time.Until(startedAt.Add(duration)).Milliseconds() < 0
}

func (cj *CronJob) RecoverRun(expRun *core.ExperimentRun) {
	defer cj.setWaitForRecovery(false)
	log.Info("recovering attack on exp run", zap.String("expRunUID", expRun.UID))
	if err := cj.recoverFunc(); err != nil {
		log.Warn("recovery failed", zap.Error(err))
	} else {
		if err := cj.scheduler.expRunStore.Update(context.Background(), expRun.UID, core.RunRecovered, ""); err != nil {
			log.Error("failed to update in DB", zap.Error(err))
		}
	}
}

func (cj *CronJob) assertNoRecovery() bool {
	cj.lock.Lock()
	waitingForRecovery := cj.waitForRecovery
	cj.lock.Unlock()
	return !waitingForRecovery
}

func (cj *CronJob) setWaitForRecovery(waitForRecovery bool) {
	cj.lock.Lock()
	cj.waitForRecovery = waitForRecovery
	cj.lock.Unlock()
}

// Run implements cron.Job interface, used when scheduling cron jobs
func (cj *CronJob) Run() {
	if !cj.assertNoRecovery() {
		log.Info("skipping scheduled execution of attack since recovery in progress", zap.String("expId", cj.experiment.Uid))
		return
	}

	var newRun *core.ExperimentRun
	var recoverTimer *time.Timer
	defer func() {
		var updErr error
		if panicRec := recover(); panicRec != nil {
			var panicErr error
			if panicErr, _ = panicRec.(error); panicErr == nil {
				panicErr = perr.New(panicRec.(string))
			}
			log.Error("scheduled run errored", zap.String("expId", cj.experiment.Uid), zap.Error(panicErr))
			if newRun != nil {
				updErr = cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunFailed, panicErr.Error())
				if recoverTimer != nil {
					cj.setWaitForRecovery(false)
					recoverTimer.Stop()
				}
			} else {
				// cannot even create a new run, maybe due to config error
				// so better to set ERROR on the experiment and remove from scheduler
				_ = cj.scheduler.Remove(cj.experiment.ID)
				updErr = cj.scheduler.expStore.Update(context.Background(), cj.experiment.Uid, core.Error, "", cj.experiment.RecoverCommand)
			}
		} else {
			log.Info("scheduled run success", zap.String("expId", cj.experiment.Uid))
			if newRun != nil {
				updErr = cj.scheduler.expRunStore.Update(context.Background(), newRun.UID, core.RunSuccess, "")
			}
		}
		if updErr != nil {
			log.Error("failed to update in DB", zap.Error(updErr))
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

	newRun = cj.experiment.NewRun()
	if err = cj.scheduler.expRunStore.NewRun(context.Background(), newRun); err != nil {
		panic(perr.WithStack(err))
	}

	if cronDuration != nil {
		cj.setWaitForRecovery(true)
		recoverTimer = time.AfterFunc(*cronDuration, func() {
			cj.RecoverRun(newRun)
		})
	}

	log.Info("executing attack on new exp run", zap.String("expRunUID", newRun.UID))
	if err = cj.attackFunc(); err != nil {
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

func (scheduler *Scheduler) Schedule(
	exp *core.Experiment, spec string, attackFunc func() error, recoverFunc func() error) error {
	cj := CronJob{
		scheduler:   scheduler,
		experiment:  exp,
		attackFunc:  attackFunc,
		recoverFunc: recoverFunc,
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
	log.Info("Starting Scheduler")
	scheduler.Cron.Start()
}
