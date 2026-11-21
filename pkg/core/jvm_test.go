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

	"github.com/onsi/gomega"
)

func TestJVMCommand(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	testCases := []struct {
		cmd    *JVMCommand
		errMsg string
	}{
		{
			&JVMCommand{},
			"pid can't be 0",
		},
		{
			&JVMCommand{
				Pid: 1234,
			},
			"action not provided",
		},
		{
			&JVMCommand{
				Pid:    1234,
				Action: "test",
			},
			"action test not supported",
		},
		{
			&JVMCommand{
				Pid:    1234,
				Action: JVMLatencyAction,
			},
			"class not provided",
		},
		{
			&JVMCommand{
				Pid:    1234,
				Action: JVMExceptionAction,
				Class:  "test",
			},
			"method not provided",
		},
		{
			&JVMCommand{
				Pid:    1234,
				Action: JVMExceptionAction,
				Class:  "test",
				Method: "test",
			},
			"",
		},
		{
			&JVMCommand{
				Pid:    1234,
				Action: JVMStressAction,
			},
			"must set one of cpu-count and mem-type",
		},
		{
			&JVMCommand{
				Pid:        1234,
				Action:     JVMStressAction,
				CPUCount:   1,
				MemoryType: "heap",
			},
			"inject stress on both CPU and memory is not support now",
		},
		{
			&JVMCommand{
				Pid:      1234,
				Action:   JVMStressAction,
				CPUCount: 1,
			},
			"",
		},
	}

	for _, testCase := range testCases {
		err := testCase.cmd.Validate()
		if len(testCase.errMsg) == 0 {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		} else {
			g.Expect(err.Error()).Should(gomega.ContainSubstring(testCase.errMsg))
		}
	}
}

func TestJVMScheduleDurationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, schedule, duration string
		invalid                  bool
	}{
		{"unscheduled", "", "", false},
		{"scheduled with duration", "@every 1h", "10s", false},
		{"scheduled invalid duration", "@every 1h", "not-a-duration", true},
		{"scheduled empty duration", "@every 1h", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := NewJVMCommand()
			command.Action = JVMGCAction
			command.Pid = 1234
			command.Schedule = tc.schedule
			command.Duration = tc.duration
			if err := command.Validate(); (err != nil) != tc.invalid {
				t.Errorf("schedule=%q duration=%q: %v", tc.schedule, tc.duration, err)
			}
		})
	}
}
