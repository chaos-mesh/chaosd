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

// TODO: support more cases such as IPv6 contained IPv4
func TestIPToCidr(t *testing.T) {
	g := NewGomegaWithT(t)
	type TestCase struct {
		name          string
		ip            string
		expectedValue string
	}
	tcs := []TestCase{
		{
			name:          "valid ipv4",
			ip:            "172.8.4.2",
			expectedValue: "172.8.4.2/32",
		},
		{
			name:          "valid ipv6",
			ip:            "2001:da8:215:4020:226:b9ff:fe2c:54f",
			expectedValue: "2001:da8:215:4020:226:b9ff:fe2c:54f/128",
		},
	}
	for _, tc := range tcs {
		g.Expect(IPToCidr(tc.ip)).To(Equal(tc.expectedValue))
	}
}

func TestResolveCidrs(t *testing.T) {
	type TestCase struct {
		name          string
		names         []string
		expectedValue []string
	}
	g := NewGomegaWithT(t)
	tcs := []TestCase{
		{
			// TODO: the last one need further discussion.
			name:          "valid names",
			names:         []string{"192.0.2.1/24", "2001:db8:a0b:12f0::1/32", "::1", "2001:da8:215:4020:226:b9ff:fe2c:54f"},
			expectedValue: []string{"192.0.2.0/24", "2001:db8::/32", "::1/128", "2001:da8:215:4020:226:b9ff:fe2c:54f/128"},
		},
	}

	for _, tc := range tcs {
		g.Expect(ResolveCidrs(tc.names)).To(Equal(tc.expectedValue))
	}
}
