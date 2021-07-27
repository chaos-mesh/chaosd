// Copyright 2020 Chaos Mesh Authors.
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

package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCheckPorts(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		ports         string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "single port",
			ports:         "2333",
			expectedValue: true,
		},
		{
			name:          "multi ports",
			ports:         "2333,2334",
			expectedValue: true,
		},
		{
			name:          "wrong port",
			ports:         "port",
			expectedValue: false,
		},
		{
			name:          "range ports",
			ports:         "2334:2335",
			expectedValue: true,
		},
		{
			name:          "wrong range ports",
			ports:         "2334:port",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(CheckPorts(tc.ports)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestCheckIPs(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		ips           string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "single ip",
			ips:           "172.16.4.4",
			expectedValue: true,
		},
		{
			name:          "multi ips",
			ips:           "172.16.4.4,172.16.4.5",
			expectedValue: true,
		},
		{
			name:          "wrong ip",
			ips:           "172.16.d.d",
			expectedValue: false,
		},
		{
			name:          "overflow ip",
			ips:           "172.16.4.256",
			expectedValue: false,
		},
		{
			name:          "rang ips",
			ips:           "172.16.4.0/16",
			expectedValue: true,
		},
		{
			name: "multi mixed ips",
			// mix rang ip and normal ip together
			ips:           "172.16.4.0/16,172.16.4.4,172.16.4.5",
			expectedValue: true,
		},
	}

	for _, tc := range tcs {
		g.Expect(CheckIPs(tc.ips)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestCheckIPProtocols(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		protocol      string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "supported protocol",
			protocol:      "tcp",
			expectedValue: true,
		},
		{
			name:          "wrong protocol",
			protocol:      "ictp",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(CheckIPProtocols(tc.protocol)).To(Equal(tc.expectedValue))
	}
}

func TestCheckPercent(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		percent       string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "valid percent",
			percent:       "1.048596",
			expectedValue: true,
		},
		{
			name:          "too big percent",
			percent:       "7355608",
			expectedValue: false,
		},
		{
			name:          "invalid percent",
			percent:       "-1/12",
			expectedValue: false,
		},
		{
			name:          "negative percent",
			percent:       "-0.618",
			expectedValue: false,
		},
	}
	for _, tc := range tcs {
		g.Expect(CheckPercent(tc.percent)).To(Equal(tc.expectedValue))
	}
}
