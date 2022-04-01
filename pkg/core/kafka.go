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

import (
	"encoding/json"
)

type KafkaAttackAction string

const (
	// Kafka actions
	KafkaFloodAction = "flood"
	KafkaIOAction    = "io"
)

var _ AttackConfig = &KafkaCommand{}

type KafkaCommand struct {
	CommonAttackConfig

	Action KafkaAttackAction
}

func (c *KafkaCommand) Validate() error {
	return nil
}

func (c *KafkaCommand) RecoverData() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *KafkaCommand) CompleteDefaults() {}

func NewKafkaCommand() *KafkaCommand {
	return &KafkaCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: KafkaAttack,
		},
	}
}
