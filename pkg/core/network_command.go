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
	"strconv"
	"strings"
	"time"

	"github.com/pingcap/errors"

	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
)

type NetworkCommand struct {
	Action      string
	Latency     string
	Jitter      string
	Correlation string
	Percent     string
	Device      string
	SourcePort  string
	EgressPort  string
	IPAddress   string
	IPProtocol  string
	Hostname    string
}

const (
	NetworkDelayAction = "delay"
	NetworkLossAction  = "loss"
)

func (n *NetworkCommand) Validate() error {
	switch n.Action {
	case NetworkDelayAction:
		return n.validNetworkDelay()
	case NetworkLossAction:
		return n.validNetworkLoss()
	default:
		return errors.Errorf("network action %s not supported", n.Action)
	}
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

	if !utils.CheckPercent(n.Correlation) {
		return errors.Errorf("correlation %s not valid", n.Correlation)
	}

	if len(n.Device) == 0 {
		return errors.New("device is required")
	}

	if !utils.CheckIPs(n.IPAddress) {
		return errors.Errorf("ip addressed %s not valid", n.IPAddress)
	}

	return checkProtocolAndPorts(n.IPProtocol, n.SourcePort, n.EgressPort)
}

func (n *NetworkCommand) validNetworkLoss() error {
	if len(n.Percent) == 0 {
		return errors.New("percent is required")
	}

	if !utils.CheckPercent(n.Percent) {
		return errors.Errorf("percent %s not valid", n.Percent)
	}

	if !utils.CheckPercent(n.Correlation) {
		return errors.Errorf("correlation %s not valid", n.Correlation)
	}

	if len(n.Device) == 0 {
		return errors.New("device is required")
	}

	if !utils.CheckIPs(n.IPAddress) {
		return errors.Errorf("ip addressed %s not valid", n.IPAddress)
	}

	return checkProtocolAndPorts(n.IPProtocol, n.SourcePort, n.EgressPort)
}

func (n *NetworkCommand) SetDefaultForNetworkDelay() {
	if len(n.Jitter) == 0 {
		n.Jitter = "0ms"
	}

	if len(n.Correlation) == 0 {
		n.Correlation = "0"
	}
}

func (n *NetworkCommand) SetDefaultForNetworkLoss() {
	if len(n.Correlation) == 0 {
		n.Correlation = "0"
	}
}

func checkProtocolAndPorts(p string, sports string, dports string) error {
	if !utils.CheckPorts(sports) {
		return errors.Errorf("source ports %s not valid", sports)
	}

	if !utils.CheckPorts(dports) {
		return errors.Errorf("egress ports %s not valid", dports)
	}

	if !utils.CheckIPProtocols(p) {
		return errors.Errorf("ip protocols %s not valid", p)
	}

	if len(sports) > 0 || len(dports) > 0 {
		if p == "tcp" || p == "udp" {
			return nil
		}

		return errors.New("ip protocol is required")
	}

	return nil
}

func (n *NetworkCommand) String() string {
	data, _ := json.Marshal(n)

	return string(data)
}

func (n *NetworkCommand) ToDelayNetem() (*pb.Netem, error) {
	delayTime, err := time.ParseDuration(n.Latency)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jitter, err := time.ParseDuration(n.Jitter)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(n.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	netem := &pb.Netem{
		Device:    n.Device,
		Time:      uint32(delayTime.Nanoseconds() / 1e3),
		DelayCorr: float32(corr),
		Jitter:    uint32(jitter.Nanoseconds() / 1e3),
	}

	return netem, nil
}

func (n *NetworkCommand) ToLossNetem() (*pb.Netem, error) {
	percent, err := strconv.ParseFloat(n.Percent, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	corr, err := strconv.ParseFloat(n.Correlation, 32)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Loss:     float32(percent),
		LossCorr: float32(corr),
	}, nil
}

func (n *NetworkCommand) ToTC(ipset string) (*pb.Tc, error) {
	tc := &pb.Tc{
		Type:       pb.Tc_NETEM,
		Ipset:      ipset,
		Protocol:   n.IPProtocol,
		SourcePort: n.SourcePort,
		EgressPort: n.EgressPort,
	}

	var (
		netem *pb.Netem
		err   error
	)
	switch n.Action {
	case NetworkDelayAction:
		if netem, err = n.ToDelayNetem(); err != nil {
			return nil, errors.WithStack(err)
		}
	case NetworkLossAction:
		if netem, err = n.ToLossNetem(); err != nil {
			return nil, errors.WithStack(err)
		}
	default:
		return nil, errors.Errorf("action %s not supported", n.Action)
	}

	tc.Netem = netem

	return tc, nil
}

func (n *NetworkCommand) ToIPSet(name string) (*pb.IPSet, error) {
	var (
		cidrs []string
		err   error
	)
	if len(n.IPAddress) > 0 {
		cidrs, err = utils.ResolveCidrs(strings.Split(n.IPAddress, ","))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if len(n.Hostname) > 0 {
		cs, err := utils.ResolveCidrs(strings.Split(n.Hostname, ","))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cidrs = append(cidrs, cs...)
	}

	return &pb.IPSet{
		Name:  name,
		Cidrs: cidrs,
	}, nil
}

func (n *NetworkCommand) NeedApplyIPSet() bool {
	if len(n.IPAddress) > 0 || len(n.Hostname) > 0 {
		return true
	}

	return false
}

func (n *NetworkCommand) NeedApplyIptables() bool {
	return false
}

func (n *NetworkCommand) NeedApplyTC() bool {
	switch n.Action {
	case NetworkDelayAction, NetworkLossAction:
		return true
	default:
		return false
	}
}

func (n *NetworkCommand) ToChain() (*pb.Chain, error) {
	return nil, nil
}
