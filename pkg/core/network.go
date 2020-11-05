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

package core

import (
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
	"github.com/pingcap/errors"
	"time"
)

type NetworkCommand struct {
	Action      string
	Latency     string
	Jitter      string
	Correlation string
	Device      string
	SourcePort  string
	EgressPort  string
	IPAddress   string
	IPProtocol  string
	Hostname    string
}

const (
	NetworkDelayAction = "delay"
)

func (n *NetworkCommand) Validate() error {
	switch n.Action {
	case NetworkDelayAction:
		return n.validNetworkDelay()
	default:
		return errors.Errorf("network action %s not supported", n.Action)
	}

	return nil
}

func (n *NetworkCommand) validNetworkDelay() error {
	if len(n.Latency) == 0 {
		return errors.New("delay is required")
	}

	if _, err := time.ParseDuration(n.Latency); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("latency %s not valid", n.Latency))
	}

	if len(n.Jitter) > 0 {
		if _, err := time.ParseDuration(n.Jitter); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("jitter %s not valid", n.Jitter))
		}
	}

	if len(n.Device) == 0 {
		return errors.New("device is required")
	}

	if !utils.CheckPorts(n.SourcePort) {
		return errors.Errorf("source ports %s not valid", n.SourcePort)
	}

	if !utils.CheckPorts(n.EgressPort) {
		return errors.Errorf("egress ports %s not valid", n.EgressPort)
	}

	if !utils.CheckIPs(n.IPAddress) {
		return errors.Errorf("ip addressed %s not valid", n.IPAddress)
	}

	if !utils.CheckIPProtocols(n.IPProtocol) {
		return errors.Errorf("ip protocols %s not valid", n.IPProtocol)
	}

	return nil
}

func (n *NetworkCommand) String() string {
	data, _ := json.Marshal(n)

	return string(data)
}
