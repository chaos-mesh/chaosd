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

	"github.com/pingcap/errors"
)

const (
	JVMPrepareType = "prepare"
	JVMAttachType  = "attach"
	JVMAgentType   = "agent"

	JVMLatencyAction   = "latency"
	JVMExceptionAction = "exception"
	JVMReturnAction    = "return"
)

type JVMCommand struct {
	// rule name, should be unique
	Name string

	// Java class
	Class string

	// the method in Java class
	Method string

	// fault action, values can be latency, exception, return
	Action string

	Value string

	// attach or agent
	Type string

	// port
	Port int

	Pid int

	Do string
}

func (j *JVMCommand) Validate() error {
	if len(j.Name) == 0 {
		return errors.New("name not provided")
	}

	if len(j.Class) == 0 {
		return errors.New("class not provided")
	}

	if len(j.Method) == 0 {
		return errors.New("method not provided")
	}

	if len(j.Action) == 0 {
		return errors.New("action not provided, action can be 'attach' or 'agent'")
	}

	if len(j.Type) == 0 {
		return errors.New("type not provided, type can be 'latency', 'exception' or 'return'")
	}

	return nil
}

func (j *JVMCommand) String() string {
	data, _ := json.Marshal(j)

	return string(data)
}
