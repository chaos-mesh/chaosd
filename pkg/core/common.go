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

import "time"

type AttackConfig interface {
	Validate() error
	Cron() string
	ScheduleDuration() (*time.Duration, error)
	// String is replacement of .Action
	String() string
	// RecoverData is replacement of earlier .String()
	RecoverData() string

	// AttackKind returns the kind of attack
	AttackKind() string

	// CompleteDefaults is used to fill flags with default values
	CompleteDefaults()
}

type SchedulerConfig struct {
	Schedule string `json:"schedule"`
	Duration string `json:"duration"`
}

func (config SchedulerConfig) Cron() string {
	return config.Schedule
}

func (config SchedulerConfig) ScheduleDuration() (*time.Duration, error) {
	if len(config.Duration) == 0 {
		return nil, nil
	}
	duration, err := time.ParseDuration(config.Duration)
	return &duration, err
}

type CommonAttackConfig struct {
	SchedulerConfig

	Action string `json:"action"`
	Kind   string `json:"kind"`
}

func (config CommonAttackConfig) String() string {
	return config.Action
}

func (config CommonAttackConfig) AttackKind() string {
	return config.Kind
}

// CompleteDefaults no-op implementation
func (config *CommonAttackConfig) CompleteDefaults() {}
