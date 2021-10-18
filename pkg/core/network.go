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

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type NetworkCommand struct {
	CommonAttackConfig

	Latency     string `json:"latency,omitempty"`
	Jitter      string `json:"jitter,omitempty"`
	Correlation string `json:"correlation,omitempty"`
	Percent     string `json:"percent,omitempty"`
	Device      string `json:"device,omitempty"`
	SourcePort  string `json:"source-port,omitempty"`
	EgressPort  string `json:"egress-port,omitempty"`
	IPAddress   string `json:"ip-address,omitempty"`
	IPProtocol  string `json:"ip-protocol,omitempty"`
	Hostname    string `json:"hostname,omitempty"`

	Direction string `json:"direction,omitempty"`

	// used for DNS attack
	DNSServer string `json:"dns-server,omitempty"`
	DNSIp     string `json:"dns-ip,omitempty"`
	DNSHost   string `json:"dns-host,omitempty"`

	// used for port occupied
	Port    string `json:"port,omitempty"`
	PortPid int32  `json:"port-pid,omitempty"`

	// only the packet which match the tcp flag can be accepted, others will be dropped.
	// only set when the IPProtocol is tcp, used for partition.
	AcceptTCPFlags string `json:"accept-tcp-flags,omitempty"`
}

var _ AttackConfig = &NetworkCommand{}

const (
	NetworkDelayAction     = "delay"
	NetworkLossAction      = "loss"
	NetworkCorruptAction   = "corrupt"
	NetworkDuplicateAction = "duplicate"
	NetworkDNSAction       = "dns"
	NetworkPartitionAction = "partition"
	NetworkPortOccupied    = "occupied"
)

func (n *NetworkCommand) Validate() error {
	if err := n.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	switch n.Action {
	case NetworkDelayAction:
		return n.validNetworkDelay()
	case NetworkLossAction, NetworkCorruptAction, NetworkDuplicateAction:
		return n.validNetworkCommon()
	case NetworkDNSAction:
		return n.validNetworkDNS()
	case NetworkPartitionAction:
		return n.validNetworkPartition()
	case NetworkPortOccupied:
		return n.validNetworkOccupied()
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

func (n *NetworkCommand) validNetworkCommon() error {
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

func (n *NetworkCommand) validNetworkPartition() error {
	if len(n.Device) == 0 {
		return errors.New("device is required")
	}

	if !utils.CheckIPs(n.IPAddress) {
		return errors.Errorf("ip addressed %s not valid", n.IPAddress)
	}

	if n.Direction != "to" && n.Direction != "from" {
		return errors.Errorf("direction should be one of 'to' and 'from'")
	}

	if len(n.AcceptTCPFlags) > 0 && n.IPProtocol != "tcp" {
		return errors.Errorf("protocol should be 'tcp' when set accept-tcp-flags")
	}

	if !utils.CheckIPProtocols(n.IPProtocol) {
		return errors.Errorf("ip protocols %s not valid", n.IPProtocol)
	}

	return nil
}

func (n *NetworkCommand) validNetworkDNS() error {
	if !utils.CheckIPs(n.DNSServer) {
		return errors.Errorf("server addresse %s not valid", n.DNSServer)
	}

	if !utils.CheckIPs(n.DNSIp) {
		return errors.Errorf("ip addresse %s not valid", n.DNSIp)
	}

	if (len(n.DNSHost) != 0 && len(n.DNSIp) == 0) || (len(n.DNSHost) == 0 && len(n.DNSIp) != 0) {
		return errors.Errorf("DNS host %s must match a DNS ip %s", n.DNSHost, n.DNSIp)
	}

	return nil
}

func (n *NetworkCommand) validNetworkOccupied() error {
	if len(n.Port) == 0 {
		return errors.New("port is required")
	}
	return nil
}

func (n *NetworkCommand) CompleteDefaults() {
	switch n.Action {
	case NetworkDelayAction:
		n.setDefaultForNetworkDelay()
	case NetworkLossAction:
		n.setDefaultForNetworkLoss()
	case NetworkDNSAction:
		n.setDefaultForNetworkDNS()
	}
}

func (n *NetworkCommand) setDefaultForNetworkDelay() {
	if len(n.Jitter) == 0 {
		n.Jitter = "0ms"
	}

	if len(n.Correlation) == 0 {
		n.Correlation = "0"
	}
}

func (n *NetworkCommand) setDefaultForNetworkLoss() {
	if len(n.Correlation) == 0 {
		n.Correlation = "0"
	}
}

func (n *NetworkCommand) setDefaultForNetworkDNS() {
	if len(n.DNSServer) == 0 {
		n.DNSServer = "123.123.123.123"
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

func (n NetworkCommand) RecoverData() string {
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
		Time:      uint32(delayTime.Nanoseconds() / 1e3),
		DelayCorr: float32(corr),
		Jitter:    uint32(jitter.Nanoseconds() / 1e3),
	}

	return netem, nil
}

func (n *NetworkCommand) ToLossNetem() (*pb.Netem, error) {
	percent, corr, err := n.parsePercentAndCorr()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Loss:     float32(percent),
		LossCorr: float32(corr),
	}, nil
}

func (n *NetworkCommand) parsePercentAndCorr() (percent float64, corr float64, err error) {
	percent, err = strconv.ParseFloat(n.Percent, 32)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	corr, err = strconv.ParseFloat(n.Correlation, 32)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}
	return
}

func (n *NetworkCommand) ToCorruptNetem() (*pb.Netem, error) {
	percent, corr, err := n.parsePercentAndCorr()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Corrupt:     float32(percent),
		CorruptCorr: float32(corr),
	}, nil
}

func (n *NetworkCommand) ToDuplicateNetem() (*pb.Netem, error) {
	percent, corr, err := n.parsePercentAndCorr()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &pb.Netem{
		Duplicate:     float32(percent),
		DuplicateCorr: float32(corr),
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
	case NetworkCorruptAction:
		if netem, err = n.ToCorruptNetem(); err != nil {
			return nil, errors.WithStack(err)
		}
	case NetworkDuplicateAction:
		if netem, err = n.ToDuplicateNetem(); err != nil {
			return nil, errors.WithStack(err)
		}
	case NetworkPartitionAction:

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
	return true
}

func (n *NetworkCommand) NeedApplyTC() bool {
	switch n.Action {
	case NetworkDelayAction, NetworkLossAction, NetworkCorruptAction, NetworkDuplicateAction:
		return true
	default:
		return false
	}
}

func (n *NetworkCommand) PartitionChain(ipset string) ([]*pb.Chain, error) {
	if n.Action != NetworkPartitionAction {
		return nil, nil
	}

	var directionStr string
	var directionChain pb.Chain_Direction
	if n.Direction == "to" {
		directionStr = "OUTPUT"
		directionChain = pb.Chain_OUTPUT
	} else if n.Direction == "from" {
		directionStr = "INPUT"
		directionChain = pb.Chain_INPUT
	} else {
		return nil, errors.New(fmt.Sprintf("direction %s not supported", n.Direction))
	}

	chains := make([]*pb.Chain, 0, 2)
	if len(n.AcceptTCPFlags) > 0 {
		chains = append(chains, &pb.Chain{
			Name:      fmt.Sprintf("%s/0", directionStr),
			Ipsets:    []string{ipset},
			Direction: directionChain,
			Protocol:  n.IPProtocol,
			TcpFlags:  n.AcceptTCPFlags,
			Target:    "ACCEPT",
		})
	}

	chains = append(chains, &pb.Chain{
		Name:      fmt.Sprintf("%s/1", directionStr),
		Ipsets:    []string{ipset},
		Direction: directionChain,
		Target:    "DROP",
	})

	return chains, nil
}

func (n *NetworkCommand) NeedApplyEtcHosts() bool {
	if len(n.DNSHost) > 0 || len(n.DNSIp) > 0 {
		return true
	}

	return false
}

func (n *NetworkCommand) NeedApplyDNSServer() bool {
	return len(n.DNSServer) > 0
}

func NewNetworkCommand() *NetworkCommand {
	return &NetworkCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: NetworkAttack,
		},
	}
}
