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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/pingcap/errors"

	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type networkAttack struct{}

var NetworkAttack AttackType = networkAttack{}

func (networkAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.NetworkCommand)
	var (
		ipsetName string
	)

	switch attack.Action {
	case core.NetworkDNSAction:
		if attack.NeedApplyEtcHosts() {
			if err = env.Chaos.applyEtcHosts(attack, env.AttackUid, env); err != nil {
				return errors.WithStack(err)
			}
		}

		if attack.NeedApplyDNSServer() {
			if err = env.Chaos.updateDNSServer(attack); err != nil {
				return errors.WithStack(err)
			}
		}

	case core.NetworkDelayAction, core.NetworkLossAction, core.NetworkCorruptAction, core.NetworkDuplicateAction:
		if attack.NeedApplyIPSet() {
			ipsetName, err = env.Chaos.applyIPSet(attack, env.AttackUid)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if attack.NeedApplyIptables() {
			if err = env.Chaos.applyIptables(attack, env.AttackUid); err != nil {
				return errors.WithStack(err)
			}
		}

		if attack.NeedApplyTC() {
			if err = env.Chaos.applyTC(attack, ipsetName, env.AttackUid); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (s *Server) applyIPSet(attack *core.NetworkCommand, uid string) (string, error) {
	ipset, err := attack.ToIPSet(fmt.Sprintf("chaos-%s", uid[:16]))
	if err != nil {
		return "", errors.WithStack(err)
	}

	if _, err := s.svr.FlushIPSets(context.Background(), &pb.IPSetsRequest{
		Ipsets:  []*pb.IPSet{ipset},
		EnterNS: false,
	}); err != nil {
		return "", errors.WithStack(err)
	}

	if err := s.ipsetRule.Set(context.Background(), &core.IPSetRule{
		Name:       ipset.Name,
		Cidrs:      strings.Join(ipset.Cidrs, ","),
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

	if newChain != nil {
		chains = append(chains, newChain)
	}

	if _, err := s.svr.SetIptablesChains(context.Background(), &pb.IptablesChainsRequest{
		Chains:  chains,
		EnterNS: false,
	}); err != nil {
		return errors.WithStack(err)
	}

	// TODO: cwen0
	//if err := s.iptablesRule.Set(context.Background(), &core.IptablesRule{
	//	Name:       newChain.Name,
	//	IPSets:     strings.Join(newChain.Ipsets, ","),
	//	Direction:  pb.Chain_Direction_name[int32(newChain.Direction)],
	//	Experiment: uid,
	//}); err != nil {
	//	return errors.WithStack(err)
	//}

	return nil
}

func (s *Server) applyTC(attack *core.NetworkCommand, ipset string, uid string) error {
	tcRules, err := s.tcRule.FindByDevice(context.Background(), attack.Device)
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
	if _, err := s.svr.SetTcs(context.Background(), &pb.TcsRequest{Tcs: tcs, Device: attack.Device, EnterNS: false}); err != nil {
		return errors.WithStack(err)
	}

	tc := &core.TcParameter{
		Device: attack.Device,
	}
	switch attack.Action {
	case core.NetworkDelayAction:
		tc.Delay = &core.DelaySpec{
			Latency:     attack.Latency,
			Correlation: attack.Correlation,
			Jitter:      attack.Jitter,
		}
	case core.NetworkLossAction:
		tc.Loss = &core.LossSpec{
			Loss:        attack.Percent,
			Correlation: attack.Correlation,
		}
	case core.NetworkCorruptAction:
		tc.Corrupt = &core.CorruptSpec{
			Corrupt:     attack.Percent,
			Correlation: attack.Correlation,
		}
	case core.NetworkDuplicateAction:
		tc.Duplicate = &core.DuplicateSpec{
			Duplicate:   attack.Percent,
			Correlation: attack.Correlation,
		}
	default:
		return errors.Errorf("network %s attack not supported", attack.Action)
	}

	tcString, err := json.Marshal(tc)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := s.tcRule.Set(context.Background(), &core.TCRule{
		Type:       pb.Tc_Type_name[int32(newTC.Type)],
		Device:     attack.Device,
		TC:         string(tcString),
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

func (s *Server) applyEtcHosts(attack *core.NetworkCommand, uid string, env Environment) error {
	recoverFlag := true
	cmd := "mv /etc/hosts /etc/hosts.chaosd." + uid + " && touch /etc/hosts"
	backupCmd := exec.Command("/bin/bash", "-c", cmd) // #nosec

	defer func() {
		if recoverFlag {
			if err := env.Chaos.recoverEtcHosts(attack, uid); err != nil {
				log.Error("Error recover env: %s\n", zap.Error(err))
			}
		}
	}()

	stdout, err := backupCmd.CombinedOutput()
	if err != nil {
		log.Error(backupCmd.String()+string(stdout), zap.Error(err))
		return errors.WithStack(err)
	}

	fileBytes, err := ioutil.ReadFile("/etc/hosts.chaosd." + uid) // #nosec
	if err != nil {
		return errors.WithStack(err)
	}

	lines := strings.Split(string(fileBytes), "\n")

	// Filter out the line of the hostname in /etc/hosts
	// example:
	// 10.86.33.102    qunarzz.com     q.qunarzz.com   common.qunarzz.com
	// 127.0.0.1       localhost
	needle := "^(\\d{1,3})(\\.\\d{1,3}){3}.*\\b" + attack.DNSHost + "\\b.*"
	re, err := regexp.Compile(needle)
	if err != nil {
		return errors.WithStack(err)
	}

	// match IP address, eg: 127.0.0.1
	reIp, err := regexp.Compile(`^(\d{1,3})(\.\d{1,3}){3}`)
	if err != nil {
		return errors.WithStack(err)
	}

	fd, err := os.OpenFile("/etc/hosts", os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			log.Error("Error closing file: %s\n", zap.Error(err))
		}
	}()

	w := bufio.NewWriter(fd)

	newFlag := true
	// if match one line, then replace it.
	for _, line := range lines {
		match := re.MatchString(line)
		if match {
			line = reIp.ReplaceAllString(line, attack.DNSIp)
			newFlag = false
		}
		line = line + "\n"
		_, err := w.WriteString(line)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	// if not match any, then add a new line.
	if newFlag {
		_, err := w.WriteString(attack.DNSIp + "\t" + attack.DNSHost + "\n")
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = w.Flush()
	if err != nil {
		return errors.WithStack(err)
	}
	err = fd.Sync()
	if err != nil {
		return errors.WithStack(err)
	}
	recoverFlag = false
	return nil
}

func (networkAttack) Recover(exp core.Experiment, env Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	attack := config.(*core.NetworkCommand)

	switch attack.Action {
	case core.NetworkDNSAction:
		if attack.NeedApplyEtcHosts() {
			if err := env.Chaos.recoverEtcHosts(attack, env.AttackUid); err != nil {
				return errors.WithStack(err)
			}
		}
		return env.Chaos.recoverDNSServer(attack)

	case core.NetworkDelayAction, core.NetworkLossAction, core.NetworkCorruptAction, core.NetworkDuplicateAction:
		if attack.NeedApplyIPSet() {
			if err := env.Chaos.recoverIPSet(env.AttackUid); err != nil {
				return errors.WithStack(err)
			}
		}

		if attack.NeedApplyIptables() {
			if err := env.Chaos.recoverIptables(env.AttackUid); err != nil {
				return errors.WithStack(err)
			}
		}

		if attack.NeedApplyTC() {
			if err := env.Chaos.recoverTC(env.AttackUid, attack.Device); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func (s *Server) recoverIPSet(uid string) error {
	if err := s.ipsetRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverIptables(uid string) error {
	if err := s.iptablesRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return errors.WithStack(err)
	}

	iptables, err := s.iptablesRule.List(context.Background())
	if err != nil {
		return errors.WithStack(err)
	}

	chains := core.IptablesRuleList(iptables).ToChains()

	if _, err := s.svr.SetIptablesChains(context.Background(), &pb.IptablesChainsRequest{
		Chains:  chains,
		EnterNS: false,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverTC(uid string, device string) error {
	if err := s.tcRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return errors.WithStack(err)
	}

	tcRules, err := s.tcRule.FindByDevice(context.Background(), device)

	tcs, err := core.TCRuleList(tcRules).ToTCs()
	if err != nil {
		return errors.WithStack(err)
	}

	if _, err := s.svr.SetTcs(context.Background(), &pb.TcsRequest{Tcs: tcs, Device: device, EnterNS: false}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) updateDNSServer(attack *core.NetworkCommand) error {
	if _, err := s.svr.SetDNSServer(context.Background(), &pb.SetDNSServerRequest{
		DnsServer: attack.DNSServer,
		Enable:    true,
		EnterNS:   false,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverDNSServer(attack *core.NetworkCommand) error {
	if _, err := s.svr.SetDNSServer(context.Background(), &pb.SetDNSServerRequest{
		Enable:  false,
		EnterNS: false,
	}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverEtcHosts(attack *core.NetworkCommand, uid string) error {
	cmd := "mv /etc/hosts.chaosd." + uid + " /etc/hosts"
	recoverCmd := exec.Command("/bin/bash", "-c", cmd) // #nosec
	stdout, err := recoverCmd.CombinedOutput()
	if err != nil {
		log.Error(recoverCmd.String()+string(stdout), zap.Error(err))
		return errors.WithStack(err)
	}
	return nil
}
