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
	Action KafkaAttackAction
	Topic  string `json:"topic,omitempty"`

	// options for fill and flood attack
	Host        string `json:"host,omitempty"`
	Port        uint16 `json:"port,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	MessageSize uint   `json:"messageSize,omitempty"`
	MaxBytes    uint64 `json:"maxBytes,omitempty"`

	// options for fill attack
	ReloadCommand string `json:"reloadCommand,omitempty"`

	// options for flood attack
	Threads uint `json:"threads,omitempty"`

	// options for fill and io attack
	ConfigFile string `json:"configFile,omitempty"`

	// options for io attack
	NonReadable bool `json:"nonReadable,omitempty"`
	NonWritable bool `json:"nonWritable,omitempty"`

	// recover data for io attack
	OriginModeOfFiles map[string]uint32 `json:"originModeOfFiles,omitempty"`
	OriginConfig      string            `json:"originConfig,omitempty"`
}

func (c *KafkaCommand) Validate() error {
	if c.Topic == "" {
		return errors.New("topic is required")
	}

	switch c.Action {
	case KafkaFillAction:
		return c.validateFillAction()
	case KafkaFloodAction:
		return c.validateFloodAction()
	case KafkaIOAction:
		return c.validateIOAction()
	default:
		return errors.Errorf("invalid action: %s", c.Action)
	}
}

func (c *KafkaCommand) validateDSNAndMessageSize() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.MessageSize == 0 {
		return errors.New("message size is required")
	}
	return nil
}

func (c *KafkaCommand) validateFillAction() error {
	if c.MaxBytes == 0 {
		return errors.New("max bytes is required")
	}
	if c.ReloadCommand == "" {
		return errors.New("reload command is required")
	}
	if _, err := os.Stat(c.ConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("config file %s not exists", c.ConfigFile)
	}
	return c.validateDSNAndMessageSize()
}

func (c *KafkaCommand) validateFloodAction() error {
	if c.Threads == 0 {
		return errors.New("threads is required")
	}
	return c.validateDSNAndMessageSize()
}

func (c *KafkaCommand) validateIOAction() error {
	if _, err := os.Stat(c.ConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("config file %s not exists", c.ConfigFile)
	}
	if !c.NonReadable && !c.NonWritable {
		return errors.New("at least one of non-readable or non-writable is required")
	}
	return nil
}

func (c *KafkaCommand) RecoverData() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *KafkaCommand) CompleteDefaults() {
	c.CommonAttackConfig.CompleteDefaults()
}

func NewKafkaCommand() *KafkaCommand {
	return &KafkaCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: KafkaAttack,
		},
		OriginModeOfFiles: make(map[string]uint32),
	}
}
