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
	"net"
	"strconv"
	"strings"
)

func CheckPorts(p string) bool {
	if len(p) == 0 {
		return true
	}

	ports := strings.Split(p, ",")
	for _, p := range ports {
		if len(p) == 0 {
			return false
		}

		ps := strings.Split(p, "-")
		if len(ps) == 0 {
			if _, err := strconv.Atoi(p); err != nil {
				return false
			}
			continue
		}

		if len(ps) != 2 {
			return false
		}

		for _, pp := range ps {
			if _, err := strconv.Atoi(pp); err != nil {
				return false
			}
		}
	}

	return true
}

func CheckIPs(i string) bool {
	if len(i) == 0 {
		return true
	}

	ips := strings.Split(i, ",")
	for _, ip := range ips {
		if ret := net.ParseIP(ip); ret == nil {
			return false
		}
	}

	return true
}

func CheckIPProtocols(p string) bool {
	if len(p) == 0 {
		return true
	}

	pros := strings.Split(p, ",")
	for _, protocol := range pros {
		if protocol != "TCP" || protocol != "UDP" || protocol != "ICMP" {
			return false
		}
	}

	return true
}
