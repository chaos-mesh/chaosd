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
	"time"

	cron "github.com/robfig/cron/v3"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type Scheduler struct {
	*cron.Cron
}

type CronJob struct {
	experiment core.Experiment
	run        func()
}

// TODO: on running, create new experiment run
// TODO: write tests for it
func (cj CronJob) Run() {
	cj.run()
}

func NewScheduler() Scheduler {
	return Scheduler{cron.New(
		cron.WithLocation(time.UTC),
		cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
	)}
}

func (scheduler Scheduler) Schedule(exp core.Experiment, spec string, task func()) error {
	cj := CronJob{experiment: exp, run: task}
	entryId, err := scheduler.AddJob(spec, cj)
	if err != nil {
		return err
	}
	cronStore.entry[exp.ID] = entryId
	return nil
}

func (scheduler Scheduler) Remove(expId uint) error {
	return nil
}
