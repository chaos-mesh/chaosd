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
	StressCPUAction = "cpu"
	StressMemAction = "mem"
)

type StressCommand struct {
	CommonAttackConfig

	Load        int
	Workers     int
	Size        string
	Options     []string
	StressngPid int32
}

var _ AttackConfig = &StressCommand{}

func (s *StressCommand) Validate() error {
	if err := s.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	if len(s.Action) == 0 {
		return errors.New("action not provided")
	}

	return nil
}

func (s StressCommand) RecoverData() string {
	data, _ := json.Marshal(s)

	return string(data)
}

func NewStressCommand() *StressCommand {
	return &StressCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: StressAttack,
		},
	}
}
