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
	RedisSentinelRestartAction = "restart"
	RedisSentinelStopAction    = "stop"
)

var _ AttackConfig = &RedisCommand{}

type RedisCommand struct {
	CommonAttackConfig

	Addr     string `json:"addr,omitempty"`
	Password string `json:"password,omitempty"`
	DB       int    `json:"db,omitempty"`
	Conf     string `json:"conf,omitempty"`
}

func (p *RedisCommand) Validate() error {
	if err := p.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	if len(p.Addr) == 0 {
		return errors.New("addr not provided")
	}
	// TODO: validate signal

	return nil
}

func (p RedisCommand) RecoverData() string {
	data, _ := json.Marshal(p)

	return string(data)
}

func NewRedisCommand() *RedisCommand {
	return &RedisCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: RedisAttack,
		},
	}
}
