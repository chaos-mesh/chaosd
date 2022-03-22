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
	// jvm action
	JVMLatencyAction   = "latency"
	JVMExceptionAction = "exception"
	JVMReturnAction    = "return"
	JVMStressAction    = "stress"
	JVMGCAction        = "gc"
	JVMRuleFileAction  = "rule-file"
	JVMRuleDataAction  = "rule-data"
	JVMMySQLAction     = "mysql"

	// for action 'mysql', 'gc' and 'stress'
	SQLHelper    = "org.chaos_mesh.byteman.helper.SQLHelper"
	GCHelper     = "org.chaos_mesh.byteman.helper.GCHelper"
	StressHelper = "org.chaos_mesh.byteman.helper.StressHelper"

	// the trigger point for 'gc' and 'stress'
	TriggerClass  = "org.chaos_mesh.chaos_agent.TriggerThread"
	TriggerMethod = "triggerFunc"

	MySQL5InjectClass  = "com.mysql.jdbc.MysqlIO"
	MySQL5InjectMethod = "sqlQueryDirect"
	MySQL5Exception    = "java.sql.SQLException(\"%s\")"

	MySQL8InjectClass  = "com.mysql.cj.NativeSession"
	MySQL8InjectMethod = "execSQL"
	MySQL8Exception    = "com.mysql.cj.exceptions.CJException(\"%s\")"
)

// byteman rule template
const (
	SimpleRuleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
AT ENTRY
IF true
DO
	{{.Do}};
ENDRULE
`

	CompleteRuleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
HELPER {{.Helper}}
AT ENTRY
BIND {{.Bind}};
IF {{.Condition}}
DO
	{{.Do}};
ENDRULE
`
)

type JVMCommand struct {
	CommonAttackConfig

	JVMCommonSpec

	JVMClassMethodSpec

	JVMStressSpec

	JVMMySQLSpec

	// rule name, should be unique, and will generate by chaosd automatically
	Name string `json:"name,omitempty"`

	// fault action, values can be latency, exception, return, stress, gc, rule-file, rule-data, mysql
	Action string `json:"action,omitempty"`

	// the return value for action 'return'
	ReturnValue string `json:"value,omitempty"`

	// the exception which needs to throw for action `exception`
	// or the exception message needs to throw in action `mysql`
	ThrowException string `json:"exception,omitempty"`

	// the latency duration for action 'latency'
	// or the latency duration in action `mysql`
	LatencyDuration int `json:"latency,omitempty"`

	// btm rule file path for action 'rule-file'
	RuleFile string `json:"rule-file,omitempty"`

	// RuleData used to save the rule file's data, will use it when recover, for action 'rule-data'
	RuleData string `json:"rule-data,omitempty"`
}

type JVMCommonSpec struct {
	// the port of agent server
	Port int `json:"port,omitempty"`

	// the pid of Java process which need to attach
	Pid int `json:"pid,omitempty"`
}

type JVMClassMethodSpec struct {
	// Java class
	Class string `json:"class,omitempty"`

	// the method in Java class
	Method string `json:"method,omitempty"`
}

type JVMStressSpec struct {
	// the CPU core number need to use, only set it when action is stress
	CPUCount int `json:"cpu-count,omitempty"`

	// the memory type need to locate, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType string `json:"mem-type,omitempty"`
}

// JVMMySQLSpec is the specification of MySQL fault injection in JVM
// only when SQL match the Database, Table and SQLType, chaosd will inject fault
// for examle:
//   SQL is "select * from test.t1",
//   only when ((Database == "test" || Database == "") && (Table == "t1" || Table == "") && (SQLType == "select" || SQLType == "")) is true, chaosd will inject fault
type JVMMySQLSpec struct {
	// the version of mysql-connector-java, only support 5.X.X(set to 5) and 8.X.X(set to 8) now
	MySQLConnectorVersion string

	// the match database
	// default value is "", means match all database
	Database string

	// the match table
	// default value is "", means match all table
	Table string

	// the match sql type
	// default value is "", means match all SQL type
	SQLType string
}

type BytemanTemplateSpec struct {
	Name      string
	Class     string
	Method    string
	Helper    string
	Bind      string
	Condition string
	Do        string

	// below is only used for stress template
	StressType      string
	StressValueName string
	StressValue     string
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

		if len(j.JVMClassMethodSpec.Method) == 0 {
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
	case JVMMySQLAction:
		if len(j.MySQLConnectorVersion) == 0 {
			return errors.New("MySQL connector version not provided")
		}
		if len(j.ThrowException) == 0 && j.LatencyDuration == 0 {
			return errors.New("must set one of exception or latency")
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
