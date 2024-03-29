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

package core

import (
	"encoding/json"

	"github.com/pingcap/errors"
)

const (
	ProcessKillAction = "kill"
	ProcessStopAction = "stop"
)

var _ AttackConfig = &ProcessCommand{}

type ProcessCommand struct {
	CommonAttackConfig

	// Process defines the process name or the process ID.
	Process    string `json:"process,omitempty"`
	Signal     int    `json:"signal,omitempty"`
	PIDs       []int
	RecoverCmd string `json:"recoverCmd,omitempty"`
	// TODO: support these feature
	// Newest       bool
	// Oldest       bool
	// Exact        bool
	// KillChildren bool
	// User         string
}

func (p *ProcessCommand) Validate() error {
	if err := p.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	if len(p.Process) == 0 {
		return errors.New("process not provided")
	}

	// TODO: validate signal

	return nil
}

func (p ProcessCommand) RecoverData() string {
	data, _ := json.Marshal(p)

	return string(data)
}

func NewProcessCommand() *ProcessCommand {
	return &ProcessCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: ProcessAttack,
		},
	}
}
