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

type AttackConfig interface {
	Validate() error
	Cron() string
	// String is replacement of .Action
	String() string
	// RecoverData is replacement of earlier .String()
	RecoverData() string
}

type SchedulerConfig struct {
	Schedule string
}

func (config SchedulerConfig) Cron() string {
	return config.Schedule
}

type CommonAttackConfig struct {
	SchedulerConfig

	Action string
}

func (config CommonAttackConfig) String() string {
	return config.Action
}
