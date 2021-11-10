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
	"fmt"

	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

const (
	JVMLatencyAction   = "latency"
	JVMExceptionAction = "exception"
	JVMReturnAction    = "return"
	JVMStressAction    = "stress"
	JVMGCAction        = "gc"
	JVMRuleFileAction  = "rule-file"
	JVMRuleDataAction  = "rule-data"
)

type JVMCommand struct {
	CommonAttackConfig

	// rule name, should be unique, and will generate by chaosd automatically
	Name string `json:"name,omitempty"`

	// Java class
	Class string `json:"class,omitempty"`

	// the method in Java class
	Method string `json:"method,omitempty"`

	// fault action, values can be latency, exception, return, stress
	Action string `json:"action,omitempty"`

	// the return value for action 'return'
	ReturnValue string `json:"value,omitempty"`

	// the exception which needs to throw for action `exception`
	ThrowException string `json:"exception,omitempty"`

	// the latency duration for action 'latency'
	LatencyDuration int `json:"latency,omitempty"`

	// the CPU core number, only set it when action is stress
	CPUCount int `json:"cpu-count,omitempty"`

	// the memory type to be located, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType string `json:"mem-type,omitempty"`

	// attach or agent
	Type string

	// the port of agent server
	Port int `json:"port,omitempty"`

	// the pid of Java process which needs to attach
	Pid int `json:"pid,omitempty"`

	// btm rule file path
	RuleFile string `json:"rule-file,omitempty"`

	// RuleData used to save the rule file's data, will use it when recover
	RuleData string `json:"rule-data,omitempty"`

	// below is only used for template
	Do string `json:"-"`

	StressType string `json:"-"`

	StressValueName string `json:"-"`

	StressValue string `json:"-"`
}

func (j *JVMCommand) Validate() error {
	if j.Pid == 0 {
		return errors.New("pid can't be 0")
	}

	switch j.Action {
	case JVMStressAction:
		if j.CPUCount == 0 && len(j.MemoryType) == 0 {
			return errors.New("must set one of cpu-count and mem-type when action is 'stress'")
		}

		if j.CPUCount > 0 && len(j.MemoryType) > 0 {
			return errors.New("inject stress on both CPU and memory is not support now")
		}
	case JVMGCAction:
		// do nothing
	case JVMExceptionAction, JVMReturnAction, JVMLatencyAction:
		if len(j.Class) == 0 {
			return errors.New("class not provided")
		}

		if len(j.Method) == 0 {
			return errors.New("method not provided")
		}
	case JVMRuleFileAction:
		if len(j.RuleFile) == 0 {
			return errors.New("rule file not provided")
		}
	case JVMRuleDataAction:
		if len(j.RuleData) == 0 {
			return errors.New("rule data not provide")
		}
	case "":
		return errors.New("action not provided")
	default:
		return errors.New(fmt.Sprintf("action %s not supported, action can be 'latency', 'exception', 'return', 'stress', 'gc', 'rule-file' of 'rule-data'", j.Action))
	}

	return nil
}

func (j *JVMCommand) RecoverData() string {
	data, _ := json.Marshal(j)

	return string(data)
}

func (j *JVMCommand) CompleteDefaults() {
	if len(j.Name) == 0 {
		j.Name = fmt.Sprintf("%s-%s-%s-%s", j.Class, j.Method, j.Action, utils.RandomStringWithCharset(5))
	}

	if j.Port == 0 {
		j.Port = 9288
	}
}

func NewJVMCommand() *JVMCommand {
	return &JVMCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: JVMAttack,
		},
	}
}
