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
	"os"

	"github.com/pkg/errors"
)

type KafkaAttackAction string

const (
	// Kafka actions
	KafkaFillAction  = "fill"
	KafkaFloodAction = "flood"
	KafkaIOAction    = "io"
)

var _ AttackConfig = &KafkaCommand{}

type KafkaCommand struct {
	CommonAttackConfig

	// global options
	Action    KafkaAttackAction
	Topic     string
	Partition int
	NoSilent  bool

	// options for fill and flood attack
	Host        string
	Port        uint16
	Username    string
	Password    string
	MessageSize uint

	// options for flood attack
	Threads          uint
	RequestPerSecond uint64

	// options for io attack
	ConfigFile  string
	NonReadable bool
	NonWritable bool

	// recover data for io attack
	PartitionDir string
}

func (c *KafkaCommand) Validate() error {
	if c.Action != KafkaFloodAction && c.Action != KafkaIOAction && c.Action != KafkaFillAction {
		return errors.Errorf("invalid action: %s", c.Action)
	}

	if c.Topic == "" {
		return errors.New("topic is required")
	}

	if c.Action == KafkaFillAction {
		if c.Host == "" {
			return errors.New("host is required")
		}

		if c.Port == 0 {
			return errors.New("port is required")
		}

		if c.MessageSize == 0 {
			return errors.New("message size is required")
		}
	}

	if c.Action == KafkaFloodAction {
		if c.Host == "" {
			return errors.New("host is required")
		}

		if c.Port == 0 {
			return errors.New("port is required")
		}

		if c.Threads == 0 {
			return errors.New("threads is required")
		}

		if c.RequestPerSecond == 0 {
			return errors.New("request per second is required")
		}

		if c.MessageSize == 0 {
			return errors.New("message size is required")
		}
	}

	if c.Action == KafkaIOAction {
		if _, err := os.Stat(c.ConfigFile); errors.Is(err, os.ErrNotExist) {
			return errors.Errorf("config file %s not exists", c.ConfigFile)
		}
		if !c.NonReadable && !c.NonWritable {
			return errors.New("at least one of non-readable or non-writable is required")
		}
	}
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
