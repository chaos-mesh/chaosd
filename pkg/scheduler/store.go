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

import cron "github.com/robfig/cron/v3"

type CronStore interface {
	Add(experimentId uint, cronEntryId cron.EntryID)
	Remove(experimentId uint) cron.EntryID
}

type cronStore struct {
	entry map[uint]cron.EntryID
}

func (cs *cronStore) Add(experimentId uint, cronEntryId cron.EntryID) {
	cs.entry[experimentId] = cronEntryId
}

func (cs *cronStore) Remove(experimentId uint) cron.EntryID {
	entryId := cs.entry[experimentId]
	delete(cs.entry, experimentId)
	return entryId
}
