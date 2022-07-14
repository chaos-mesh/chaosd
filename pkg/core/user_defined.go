// Copyright 2022 Chaos Mesh Authors.
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

var _ AttackConfig = &UserDefinedOption{}

type UserDefinedOption struct {
	CommonAttackConfig

	AttackCmd  string `json:"attackCmd,omitempty"`
	RecoverCmd string `json:"recoverCmd,omitempty"`
}

func (u *UserDefinedOption) Validate() error {
	if err := u.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	if len(u.AttackCmd) == 0 {
		return errors.New("attack command not provided")
	}

	if len(u.RecoverCmd) == 0 {
		return errors.New("recover command not provided")
	}

	return nil
}

func (u *UserDefinedOption) RecoverData() string {
	data, _ := json.Marshal(u)

	return string(data)
}

func NewUserDefinedOption() *UserDefinedOption {
	return &UserDefinedOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: UserDefinedAttack,
		},
	}
}
