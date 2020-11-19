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

package chaosd

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
)

const (
	NetworkAttack = "network attack"
)

func (s *Server) NetworkAttack(attack *core.NetworkCommand) (string, error) {
	var (
		ipsetName string
		err       error
	)
	uid := uuid.New().String()

	if attack.NeedApplyIPSet() {
		ipsetName, err = s.applyIPSet(attack, uid)
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	if attack.NeedApplyIptables() {
		if err := s.applyIptables(attack, uid); err != nil {
			return "", errors.WithStack(err)
		}
	}

	if attack.NeedApplyTC() {
		if err := s.applyTC(attack, ipsetName, uid); err != nil {
			return "", errors.WithStack(err)
		}
	}

	if err := s.exp.Update(context.Background(), uid, core.Success, "", attack.String()); err != nil {
		return "", errors.WithStack(err)
	}

	return uid, nil
}

func (s *Server) applyIPSet(attack *core.NetworkCommand, uid string) (string, error) {
	ipset, err := attack.ToIPSet(fmt.Sprintf("chaos-%s", uid[:16]))
	if err != nil {
		return "", errors.WithStack(err)
	}

	if err := flushIPSet(context.Background(), "", ipset); err != nil {
		return "", errors.WithStack(err)
	}
	if err := s.ipsetRule.Set(context.Background(), &core.IPSetRule{
		Name:       ipset.Name,
		Cidrs:      ipset.Cidrs,
		Experiment: uid,
	}); err != nil {
		return "", errors.WithStack(err)
	}

	return ipset.Name, nil
}

func (s *Server) applyIptables(attack *core.NetworkCommand, uid string) error {
	iptables, err := s.iptablesRule.List(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}
	chains := core.IptablesRuleList(iptables).ToChains()
	newChain, err := attack.ToChain()
	if err != nil {
		return errors.WithStack(err)
	}

	chains = append(chains, newChain)
	if err := s.SetNodeIptablesChains(context.Background(), chains); err != nil {
		return errors.WithStack(err)
	}

	if err := s.iptablesRule.Set(context.Background(), &core.IptablesRule{
		Name:       newChain.Name,
		IPSets:     newChain.Ipsets,
		Direction:  pb.Chain_Direction_name[int32(newChain.Direction)],
		Experiment: uid,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) applyTC(attack *core.NetworkCommand, ipset string, uid string) error {
	tcRules, err := s.tcRule.List(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}

	tcs, err := core.TCRuleList(tcRules).ToTCs()
	if err != nil {
		return errors.WithStack(err)
	}

	newTC, err := attack.ToTC(ipset)
	if err != nil {
		return errors.WithStack(err)
	}

	tcs = append(tcs, newTC)
	if err := s.SetNodeTcRules(context.Background(), &pb.TcsRequest{Tcs: tcs}); err != nil {
		return errors.WithStack(err)
	}

	if err := s.tcRule.Set(context.Background(), &core.TCRule{
		Type: pb.Tc_Type_name[int32(newTC.Type)],
		TcParameter: core.TcParameter{
			Device: attack.Device,
			Delay: &core.DelaySpec{
				Latency:     attack.Latency,
				Correlation: attack.Correlation,
				Jitter:      attack.Latency,
			},
		},
		IPSet:      newTC.Ipset,
		Protocal:   newTC.Protocol,
		SourcePort: newTC.SourcePort,
		EgressPort: newTC.EgressPort,
		Experiment: uid,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) RecoverNetworkAttack(uid string, attack *core.NetworkCommand) error {
	switch attack.Action {
	case core.NetworkDelayAction:
		if err := s.SetNodeTcRules(context.Background(), &pb.TcsRequest{}); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := s.exp.Update(context.Background(), uid, core.Destroyed, "", attack.String()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
