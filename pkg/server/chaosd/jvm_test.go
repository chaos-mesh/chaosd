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

package chaosd

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func TestGenerateRuleData(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		cmd      *core.JVMCommand
		ruleData string
	}{
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMExceptionAction,
				JVMClassMethodSpec: core.JVMClassMethodSpec{
					Class:  "testClass",
					Method: "testMethod",
				},
				ThrowException: "java.io.IOException(\"BOOM\")",
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\tthrow new java.io.IOException(\"BOOM\");\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMReturnAction,
				JVMClassMethodSpec: core.JVMClassMethodSpec{
					Class:  "testClass",
					Method: "testMethod",
				},
				ReturnValue: "\"test\"",
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\treturn \"test\";\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMLatencyAction,
				JVMClassMethodSpec: core.JVMClassMethodSpec{
					Class:  "testClass",
					Method: "testMethod",
				},
				LatencyDuration: 5000,
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\tThread.sleep(5000);\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMStressAction,
				JVMStressSpec: core.JVMStressSpec{
					CPUCount: 1,
				},
			},
			"\nRULE test\nSTRESS CPU\nCPUCOUNT 1\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMStressAction,
				JVMStressSpec: core.JVMStressSpec{
					MemoryType: "heap",
				},
			},
			"\nRULE test\nSTRESS MEMORY\nMEMORYTYPE heap\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMMySQLAction,
				JVMMySQLSpec: core.JVMMySQLSpec{
					MySQLConnectorVersion: "8",
					Database:              "test",
					Table:                 "t1",
					SQLType:               "select",
				},
				ThrowException: "BOOM",
			},
			"\nRULE test\nCLASS com.mysql.cj.NativeSession\nMETHOD execSQL\nHELPER org.chaos_mesh.byteman.helper.SQLHelper\nAT ENTRY\nBIND flag:boolean=matchDBTable(\"\", $2, \"test\", \"t1\", \"select\");\nIF flag\nDO\n\tthrow new com.mysql.cj.exceptions.CJException(BOOM);\nENDRULE\n",
		},
		{
			&core.JVMCommand{
				Name: "test",
				JVMCommonSpec: core.JVMCommonSpec{
					Pid: 1234,
				},
				Action: core.JVMMySQLAction,
				JVMMySQLSpec: core.JVMMySQLSpec{
					MySQLConnectorVersion: "8",
					Database:              "test",
					Table:                 "t1",
					SQLType:               "select",
				},
				LatencyDuration: 5000,
			},
			"\nRULE test\nCLASS com.mysql.cj.NativeSession\nMETHOD execSQL\nHELPER org.chaos_mesh.byteman.helper.SQLHelper\nAT ENTRY\nBIND flag:boolean=matchDBTable(\"\", $2, \"test\", \"t1\", \"select\");\nIF flag\nDO\n\tThread.sleep(5000);\nENDRULE\n",
		},
	}

	for _, testCase := range testCases {
		ruleData, err := generateRuleData(testCase.cmd)
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(ruleData).Should(Equal(testCase.ruleData))
	}
}
