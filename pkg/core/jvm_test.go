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
	"testing"

	. "github.com/onsi/gomega"
)

func TestJVMCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		cmd    *JVMCommand
		errMsg string
	}{
		{
			&JVMCommand{},
			"type not provided",
		},
		{
			&JVMCommand{
				Type: JVMInstallType,
			},
			"pid can't be 0",
		},
		{
			&JVMCommand{
				Type: JVMInstallType,
				Pid:  123,
			},
			"",
		},
		{
			&JVMCommand{
				Type: JVMSubmitType,
			},
			"action not provided",
		},
		{
			&JVMCommand{
				Type:   JVMSubmitType,
				Action: "test",
			},
			"action test not supported",
		},
		{
			&JVMCommand{
				Type:   JVMSubmitType,
				Action: JVMLatencyAction,
			},
			"class not provided",
		},
		{
			&JVMCommand{
				Type:   JVMSubmitType,
				Action: JVMExceptionAction,
				Class:  "test",
			},
			"method not provided",
		},
		{
			&JVMCommand{
				Type:   JVMSubmitType,
				Action: JVMExceptionAction,
				Class:  "test",
				Method: "test",
			},
			"",
		},
		{
			&JVMCommand{
				Type:   JVMSubmitType,
				Action: JVMStressAction,
			},
			"must set one of cpu-count and mem-size",
		},
		{
			&JVMCommand{
				Type:       JVMSubmitType,
				Action:     JVMStressAction,
				CPUCount:   1,
				MemorySize: 1,
			},
			"inject stress on both CPU and memory is not support now",
		},
		{
			&JVMCommand{
				Type:     JVMSubmitType,
				Action:   JVMStressAction,
				CPUCount: 1,
			},
			"",
		},
	}

	for _, testCase := range testCases {
		err := testCase.cmd.Validate()
		if len(testCase.errMsg) == 0 {
			g.Expect(err).ShouldNot(HaveOccurred())
		} else {
			g.Expect(err.Error()).Should(ContainSubstring(testCase.errMsg))
		}
	}
}
