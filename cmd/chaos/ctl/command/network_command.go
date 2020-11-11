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

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

func NewNetworkAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network <subcommand>",
		Short: "Network attack related commands",
	}

	cmd.AddCommand(
		NewNetworkDelayCommand(),
	)

	return cmd
}

var nFlag core.NetworkCommand

func NewNetworkDelayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delay [option]",
		Short: "delay network",

		Run: networkDelayCommandFunc,
	}

	cmd.Flags().StringVarP(&nFlag.Latency, "latency", "l", "",
		"delay egress time, time units: ns, us (or µs), ms, s, m, h.")
	cmd.Flags().StringVarP(&nFlag.Jitter, "jitter", "j", "",
		"jitter time, time units: ns, us (or µs), ms, s, m, h.")
	cmd.Flags().StringVarP(&nFlag.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&nFlag.EgressPort, "egress-port", "e", "",
		"only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&nFlag.SourcePort, "source-port", "s", "",
		"only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&nFlag.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&nFlag.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&nFlag.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")
	nFlag.Action = core.NetworkDelayAction
	nFlag.SetDefault()

	return cmd
}

func networkDelayCommandFunc(cmd *cobra.Command, args []string) {
	if err := nFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.NetworkAttack(&nFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack network successfully, uid: %s", uid))
}
