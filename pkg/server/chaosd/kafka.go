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

package chaosd

import "github.com/chaos-mesh/chaosd/pkg/core"

type kafkaAttack struct{}

var KafkaAttack AttackType = kafkaAttack{}

func (j kafkaAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	return nil
}

func (j kafkaAttack) Recover(exp core.Experiment, env Environment) error {
	return nil
}
