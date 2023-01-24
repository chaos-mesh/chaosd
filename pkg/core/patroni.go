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
	SwitchoverAction = "switchover"
	FailoverAction   = "failover"
)

var _ AttackConfig = &PatroniCommand{}

type PatroniCommand struct {
	CommonAttackConfig

	Address      string `json:"address,omitempty"`
	Candidate    string `json:"candidate,omitempty"`
	Leader       string `json:"leader,omitempty"`
	User         string `json:"user,omitempty"`
	Password     string `json:"password,omitempty"`
	Scheduled_at string `json:"scheduled_at,omitempty"`
	RecoverCmd   string `json:"recoverCmd,omitempty"`
}

func (p *PatroniCommand) Validate() error {
	if err := p.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	if len(p.Address) == 0 {
		return errors.New("address not provided")
	}

	if len(p.User) == 0 {
		return errors.New("patroni user not provided")
	}

	if len(p.Password) == 0 {
		return errors.New("patroni password not provided")
	}

	return nil
}

func (p PatroniCommand) RecoverData() string {
	data, _ := json.Marshal(p)

	return string(data)
}

func NewPatroniCommand() *PatroniCommand {
	return &PatroniCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: PatroniAttack,
		},
	}
}
