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
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/go-logr/zapr"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	perrors "github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"

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
				return perrors.WithStack(err)
			}
		}

		if attack.NeedApplyDNSServer() {
			if err = env.Chaos.updateDNSServer(attack); err != nil {
				return perrors.WithStack(err)
			}
		}

	case core.NetworkPortOccupiedAction:
		return env.Chaos.applyPortOccupied(attack)

	case core.NetworkDelayAction, core.NetworkLossAction, core.NetworkCorruptAction, core.NetworkDuplicateAction, core.NetworkBandwidthAction, core.NetworkPartitionAction:
		if attack.NeedApplyIPSet() {
			ipsetName, err = env.Chaos.applyIPSet(attack, env.AttackUid)
			if err != nil {
				return perrors.WithStack(err)
			}
		}

		if attack.NeedApplyIptables() {
			if err = env.Chaos.applyIptables(attack, ipsetName, env.AttackUid); err != nil {
				return perrors.WithStack(err)
			}
		}

		if attack.NeedApplyTC() {
			if err = env.Chaos.applyTC(attack, ipsetName, env.AttackUid); err != nil {
				return perrors.WithStack(err)
			}
		}

	case core.NetworkNICDownAction:
		if err := env.Chaos.getNICIP(attack); err != nil {
			return perrors.WithStack(err)
		}

		NICDownCommand := fmt.Sprintf("ifconfig %s down", attack.Device)

		cmd := exec.Command("bash", "-c", NICDownCommand)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return perrors.WithStack(err)
		}

		if attack.Duration != "-1" {
			err := env.Chaos.recoverNICDownScheduled(attack)
			return perrors.WithStack(err)
		}
	case core.NetworkFloodAction:
		return env.Chaos.applyFlood(attack)
	}

	return nil
}

func (s *Server) applyIPSet(attack *core.NetworkCommand, uid string) (string, error) {
	ipset, err := attack.ToIPSet(fmt.Sprintf("chaos-%.16s", uid))
	if err != nil {
		return "", perrors.WithStack(err)
	}

	if _, err := s.svr.FlushIPSets(context.Background(), &pb.IPSetsRequest{
		Ipsets:  []*pb.IPSet{ipset},
		EnterNS: false,
	}); err != nil {
		return "", perrors.WithStack(err)
	}

	if err := s.ipsetRule.Set(context.Background(), &core.IPSetRule{
		Name:       ipset.Name,
		Cidrs:      strings.Join(ipset.Cidrs, ","),
		Experiment: uid,
	}); err != nil {
		return "", perrors.WithStack(err)
	}

	return ipset.Name, nil
}

func (s *Server) applyIptables(attack *core.NetworkCommand, ipset, uid string) error {
	iptables, err := s.iptablesRule.List(context.Background())
	if err != nil {
		return perrors.WithStack(err)
	}
	chains := core.IptablesRuleList(iptables).ToChains()
	// Presently, only partition and delay with `accept-tcp-flags` need to add additional chains
	if attack.NeedAdditionalChains() {
		newChains, err := attack.AdditionalChain(ipset)
		if err != nil {
			return perrors.WithStack(err)
		}
		chains = append(chains, newChains...)
	}

	if _, err := s.svr.SetIptablesChains(context.Background(), &pb.IptablesChainsRequest{
		Chains:  chains,
		EnterNS: false,
	}); err != nil {
		return perrors.WithStack(err)
	}

	// TODO: cwen0
	//if err := s.iptablesRule.Set(context.Background(), &core.IptablesRule{
	//	Name:       newChain.Name,
	//	IPSets:     strings.Join(newChain.Ipsets, ","),
	//	Direction:  pb.Chain_Direction_name[int32(newChain.Direction)],
	//	Experiment: uid,
	//}); err != nil {
	//	return perrors.WithStack(err)
	//}

	return nil
}

func (s *Server) applyTC(attack *core.NetworkCommand, ipset string, uid string) error {
	tcRules, err := s.tcRule.FindByDevice(context.Background(), attack.Device)
	if err != nil {
		return perrors.WithStack(err)
	}

	tcs, err := core.TCRuleList(tcRules).ToTCs()
	if err != nil {
		return perrors.WithStack(err)
	}

	newTC, err := attack.ToTC(ipset)
	if err != nil {
		return perrors.WithStack(err)
	}

	tcs = append(tcs, newTC)

	if _, err := s.svr.SetTcs(context.Background(), &pb.TcsRequest{Tcs: tcs, EnterNS: false}); err != nil {
		return perrors.WithStack(err)
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
	case core.NetworkBandwidthAction:
		tc.Bandwidth = &core.BandwidthSpec{
			Rate:     attack.Rate,
			Limit:    attack.Limit,
			Buffer:   attack.Buffer,
			Peakrate: attack.Peakrate,
			Minburst: attack.Minburst,
		}
	default:
		return perrors.Errorf("network %s attack not supported", attack.Action)
	}

	tcString, err := json.Marshal(tc)
	if err != nil {
		return perrors.WithStack(err)
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
		return perrors.WithStack(err)
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
		return perrors.WithStack(err)
	}

	fileBytes, err := ioutil.ReadFile("/etc/hosts.chaosd." + uid) // #nosec
	if err != nil {
		return perrors.WithStack(err)
	}

	lines := strings.Split(string(fileBytes), "\n")

	// Filter out the line of the hostname in /etc/hosts
	// example:
	// 10.86.33.102    qunarzz.com     q.qunarzz.com   common.qunarzz.com
	// 127.0.0.1       localhost
	needle := "^(\\d{1,3})(\\.\\d{1,3}){3}.*\\b" + attack.DNSDomainName + "\\b.*"
	re, err := regexp.Compile(needle)
	if err != nil {
		return perrors.WithStack(err)
	}

	// match IP address, eg: 127.0.0.1
	reIp, err := regexp.Compile(`^(\d{1,3})(\.\d{1,3}){3}`)
	if err != nil {
		return perrors.WithStack(err)
	}

	fd, err := os.OpenFile("/etc/hosts", os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return perrors.WithStack(err)
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
			return perrors.WithStack(err)
		}
	}
	// if not match any, then add a new line.
	if newFlag {
		_, err := w.WriteString(attack.DNSIp + "\t" + attack.DNSDomainName + "\n")
		if err != nil {
			return perrors.WithStack(err)
		}
	}

	err = w.Flush()
	if err != nil {
		return perrors.WithStack(err)
	}
	err = fd.Sync()
	if err != nil {
		return perrors.WithStack(err)
	}
	recoverFlag = false
	return nil
}

func (s *Server) applyFlood(attack *core.NetworkCommand) error {
	cmd := bpm.DefaultProcessBuilder("bash", "-c", fmt.Sprintf("iperf -u -c %s -t %s -p %s -P %d -b %s", attack.IPAddress, attack.Duration, attack.Port, attack.Parallel, attack.Rate)).
		Build(context.Background())

	// Build will set SysProcAttr.Pdeathsig = syscall.SIGTERM, and so iperf will exit while chaosd exit
	// so reset it here
	cmd.Cmd.SysProcAttr = &syscall.SysProcAttr{}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger := zapr.NewLogger(zapLogger)
	backgroundProcessManager := bpm.StartBackgroundProcessManager(nil, logger)
	_, err = backgroundProcessManager.StartProcess(context.Background(), cmd)
	if err != nil {
		return err
	}

	attack.IperfPid = int32(cmd.Process.Pid)
	log.Info("Start iperf process successfully", zap.String("command", cmd.String()), zap.Int32("Pid", attack.IperfPid))

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
				return perrors.WithStack(err)
			}
		}
		return env.Chaos.recoverDNSServer(attack)
	case core.NetworkPortOccupiedAction:
		return env.Chaos.recoverPortOccupied(attack, env.AttackUid)
	case core.NetworkDelayAction, core.NetworkLossAction, core.NetworkCorruptAction, core.NetworkDuplicateAction, core.NetworkPartitionAction, core.NetworkBandwidthAction:
		if attack.NeedApplyIPSet() {
			if err := env.Chaos.recoverIPSet(env.AttackUid); err != nil {
				return perrors.WithStack(err)
			}
		}

		if attack.NeedApplyIptables() {
			if err := env.Chaos.recoverIptables(env.AttackUid); err != nil {
				return perrors.WithStack(err)
			}
		}

		if attack.NeedApplyTC() {
			if err := env.Chaos.recoverTC(env.AttackUid, attack.Device); err != nil {
				return perrors.WithStack(err)
			}
		}
	case core.NetworkNICDownAction:
		return env.Chaos.recoverNICDown(attack)
	case core.NetworkFloodAction:
		return env.Chaos.recoverFlood(attack)
	}
	return nil
}

func (s *Server) recoverIPSet(uid string) error {
	if err := s.ipsetRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverIptables(uid string) error {
	if err := s.iptablesRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return perrors.WithStack(err)
	}

	iptables, err := s.iptablesRule.List(context.Background())
	if err != nil {
		return perrors.WithStack(err)
	}

	chains := core.IptablesRuleList(iptables).ToChains()

	if _, err := s.svr.SetIptablesChains(context.Background(), &pb.IptablesChainsRequest{
		Chains:  chains,
		EnterNS: false,
	}); err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverTC(uid string, device string) error {
	if err := s.tcRule.DeleteByExperiment(context.Background(), uid); err != nil {
		return perrors.WithStack(err)
	}

	tcRules, err := s.tcRule.FindByDevice(context.Background(), device)

	tcs, err := core.TCRuleList(tcRules).ToTCs()
	if err != nil {
		return perrors.WithStack(err)
	}

	if _, err := s.svr.SetTcs(context.Background(), &pb.TcsRequest{Tcs: tcs, EnterNS: false}); err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) updateDNSServer(attack *core.NetworkCommand) error {
	if _, err := s.svr.SetDNSServer(context.Background(), &pb.SetDNSServerRequest{
		DnsServer: attack.DNSServer,
		Enable:    true,
		EnterNS:   false,
	}); err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverDNSServer(attack *core.NetworkCommand) error {
	if _, err := s.svr.SetDNSServer(context.Background(), &pb.SetDNSServerRequest{
		Enable:  false,
		EnterNS: false,
	}); err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) applyPortOccupied(attack *core.NetworkCommand) error {

	if len(attack.Port) == 0 {
		return nil
	}

	flag, err := checkPortIsListened(attack.Port)
	if err != nil {
		if flag {
			return perrors.Errorf("port %s has been occupied", attack.Port)
		}
		return perrors.WithStack(err)
	}

	if flag {
		return perrors.Errorf("port %s has been occupied", attack.Port)
	}

	args := fmt.Sprintf("-p=%s", attack.Port)
	cmd := bpm.DefaultProcessBuilder("PortOccupyTool", args).Build(context.Background())

	cmd.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger := zapr.NewLogger(zapLogger)
	backgroundProcessManager := bpm.StartBackgroundProcessManager(nil, logger)
	_, err = backgroundProcessManager.StartProcess(context.Background(), cmd)
	if err != nil {
		return perrors.WithStack(err)
	}

	attack.PortPid = int32(cmd.Process.Pid)

	return nil
}

func checkPortIsListened(port string) (bool, error) {
	checkStatement := fmt.Sprintf("lsof -i:%s | awk '{print $2}' | grep -v PID", port)
	cmd := exec.Command("sh", "-c", checkStatement)

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if err.Error() == "exit status 1" && string(stdout) == "" {
			return false, nil
		}
		log.Error(cmd.String()+string(stdout), zap.Error(err))
		return true, perrors.WithStack(err)
	}

	if string(stdout) == "" {
		return false, nil
	}
	return true, nil
}

func (s *Server) recoverPortOccupied(attack *core.NetworkCommand, uid string) error {

	proc, err := process.NewProcess(attack.PortPid)
	if err != nil {
		return err
	}

	procName, err := proc.Name()
	if err != nil {
		return err
	}

	if !strings.Contains(procName, "PortOccupyTool") {
		log.Warn("the process is not PortOccupyTool, maybe it is killed by manual")
		return nil
	}

	if err := proc.Kill(); err != nil {
		log.Error("the port occupy process kill failed", zap.Error(err))
		return err
	}
	return nil
}

func (s *Server) recoverEtcHosts(attack *core.NetworkCommand, uid string) error {
	cmd := "mv /etc/hosts.chaosd." + uid + " /etc/hosts"
	recoverCmd := exec.Command("/bin/bash", "-c", cmd) // #nosec
	stdout, err := recoverCmd.CombinedOutput()
	if err != nil {
		log.Error(recoverCmd.String()+string(stdout), zap.Error(err))
		return perrors.WithStack(err)
	}
	return nil
}

func (s *Server) recoverNICDown(attack *core.NetworkCommand) error {
	NICUpCommand := fmt.Sprintf("ifconfig %s %s up", attack.Device, attack.IPAddress)

	recoverCmd := exec.Command("bash", "-c", NICUpCommand)
	_, err := recoverCmd.CombinedOutput()
	if err != nil {
		return perrors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverNICDownScheduled(attack *core.NetworkCommand) error {
	NICUpCommand := fmt.Sprintf("sleep %s && ifconfig %s %s up", attack.Duration, attack.Device, attack.IPAddress)

	recoverCmd := exec.Command("bash", "-c", NICUpCommand)
	_, err := recoverCmd.CombinedOutput()
	if err != nil {
		return perrors.WithStack(err)
	}
	return nil
}

func (s *Server) recoverFlood(attack *core.NetworkCommand) error {
	proc, err := process.NewProcess(attack.IperfPid)
	if err != nil {
		if errors.Is(err, process.ErrorProcessNotRunning) || errors.Is(err, fs.ErrNotExist) {
			log.Warn("Failed to get iperf process", zap.Error(err))
			return nil
		}

		return err
	}

	procName, err := proc.Name()
	if err != nil {
		return err
	}

	if !strings.Contains(procName, "iperf") {
		log.Warn("the process is not iperf, maybe it is killed by manual")
		return nil
	}

	if err := proc.Kill(); err != nil {
		log.Error("the iperf process kill failed", zap.Error(err))
		return err
	}

	return nil
}

// getNICIP() uses `ifconfig` to get interfaces' IP. The reason for
// not using net.Interfaces() is that net.Interfaces() can't get
// sub interfaces.
func (s *Server) getNICIP(attack *core.NetworkCommand) error {
	getIPCommand := fmt.Sprintf("ifconfig %s | awk '/inet\\>/ {print $2}'", attack.Device)

	cmd := exec.Command("bash", "-c", getIPCommand)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return perrors.WithStack(err)
	}

	if err = cmd.Start(); err != nil {
		return perrors.WithStack(err)
	}

	stdoutBytes := make([]byte, 1024)
	_, err = stdout.Read(stdoutBytes)
	if err != nil {
		return perrors.WithStack(err)
	}
	// When stdoutBytes is converted to string, the string will be IPAddress with a few unnecessary
	// zeros, which makes IPAddress' format wrong, so the trailing zeros needs to be trimmed.
	attack.IPAddress = strings.TrimRight(string(stdoutBytes), "\n\x00")

	return nil
}
