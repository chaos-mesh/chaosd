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
	JVMInstallType = "install"
	JVMSubmitType  = "submit"

	JVMLatencyAction   = "latency"
	JVMExceptionAction = "exception"
	JVMReturnAction    = "return"
	JVMStressAction    = "stress"
	JVMGCAction        = "gc"
	JVMRuleFileAction  = "rule_file"
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
	LatencyDuration string `json:"latency,omitempty"`

	// the CPU core number need to use, only set it when action is stress
	CPUCount int `json:"cpu-count,omitempty"`

	// the memory type need to locate, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType string `json:"mem-type,omitempty"`

	// attach or agent
	Type string

	// the port of agent server
	Port int `json:"port,omitempty"`

	// the pid of Java process which need to attach
	Pid int `json:"pid,omitempty"`

	// btm rule file path
	RuleFile string `json:"rule-file,omitempty"`

	// RuleData used to save the rule file's data, will use it when recover
	RuleData []byte `json:"rule-data,omitempty"`

	// below is only used for template
	Do string

	StressType string

	StressValueName string

	StressValue string
}

func (j *JVMCommand) Validate() error {
	switch j.Type {
	case JVMInstallType:
		if j.Pid == 0 {
			return errors.New("pid can't be 0")
		}
	case JVMSubmitType:
		switch j.Action {
		case JVMStressAction:
			if j.CPUCount == 0 && len(j.MemoryType) == 0 {
				return errors.New("must set one of cpu-count and mem-size when action is 'stress'")
			}

			if j.CPUCount > 0 && len(j.MemoryType) > 0 {
				return errors.New("inject stress on both CPU and memory is not support now")
			}

			if len(j.MemoryType) > 0 {
				if j.MemoryType != "heap" && j.MemoryType != "stack" {
					return errors.New("memory type should be one of 'heap' and 'stack'")
				}
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
		case "":
			return errors.New("action not provided, action can be 'latency', 'exception', 'return', 'stress' or 'gc'")
		default:
			return errors.New(fmt.Sprintf("action %s not supported, action can be 'latency', 'exception', 'return', 'stress' or 'gc'", j.Action))
		}

	case "":
		return errors.New("type not provided, type can be 'install' or 'submit'")
	default:
		return errors.New(fmt.Sprintf("type %s not supported, type can be 'install' or 'submit'", j.Type))
	}

	return nil
}

func (j *JVMCommand) RecoverData() string {
	data, _ := json.Marshal(j)

	return string(data)
}

func (j *JVMCommand) CompleteDefaults() {
	if j.Type == JVMSubmitType {
		if len(j.Name) == 0 {
			j.Name = fmt.Sprintf("%s-%s-%s-%s-%s", j.Class, j.Method, j.Action, j.Type, utils.RandomStringWithCharset(5))
		}
	}
}

func NewJVMCommand() *JVMCommand {
	return &JVMCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: JVMAttack,
		},
	}
}
