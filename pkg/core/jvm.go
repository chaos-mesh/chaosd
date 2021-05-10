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
)

type JVMCommand struct {
	CommonAttackConfig

	// rule name, should be unique, and will generate by chaosd automatically
	Name string

	// Java class
	Class string

	// the method in Java class
	Method string

	// fault action, values can be latency, exception, return, stress
	Action string

	// the return value for action 'return'
	ReturnValue string

	// the exception which needs to throw dor action `exception`
	ThrowException string

	// the latency duration for action 'latency'
	LatencyDuration string

	// the CPU core number need to use, only set it when action is stress
	CPUCount int

	// the memory size need to locate, only set it when action is stress
	MemorySize int

	// attach or agent
	Type string

	// the port of agent server
	Port int

	// the pid of Java process which need to attach
	Pid int

	// below is only used for template
	Do string

	StressType string

	StressValueName string

	StressValue int
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
			if j.CPUCount == 0 && j.MemorySize == 0 {
				return errors.New("must set one of cpu-count and mem-size when action is 'stress'")
			}

			if j.CPUCount > 0 && j.MemorySize > 0 {
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
		case "":
			return errors.New("action not provided, action can be 'latency', 'exception', 'return', 'stress' or 'gc'")
		default:
			return errors.New(fmt.Sprintf("action %s not supported, action can be 'latency', 'exception', 'return', 'stress' or 'gc'", j.Action))
		}

		if len(j.Name) == 0 {
			j.Name = fmt.Sprintf("%s-%s-%s-%s-%s", j.Class, j.Method, j.Action, j.Type, utils.RandomStringWithCharset(5))
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

func NewJVMCommand() *JVMCommand {
	return &JVMCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: JVMAttack,
		},
	}
}
