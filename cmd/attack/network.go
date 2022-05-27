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

package attack

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func NewNetworkAttackCommand(uid *string) *cobra.Command {
	options := core.NewNetworkCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.NetworkCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "network <subcommand>",
		Short: "Network attack related commands",
	}

	cmd.AddCommand(
		NewNetworkDelayCommand(dep, options),
		NewNetworkLossCommand(dep, options),
		NewNetworkCorruptCommand(dep, options),
		NetworkDuplicateCommand(dep, options),
		NetworkPartitionCommand(dep, options),
		NetworkDNSCommand(dep, options),
		NewNetworkPortOccupiedCommand(dep, options),
		NewNetworkBandwidthCommand(dep, options),
		NewNICDownCommand(dep, options),
		NewNetworkFloodCommand(dep, options),
	)

	return cmd
}

func NewNetworkDelayCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delay",
		Short: "delay network",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDelayAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Latency, "latency", "l", "",
		"delay egress time, time units: ns, us (or µs), ms, s, m, h.")
	cmd.Flags().StringVarP(&options.Jitter, "jitter", "j", "",
		"jitter time, time units: ns, us (or µs), ms, s, m, h.")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "correlation is percentage (10 is 10%)")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")

	return cmd
}

func NewNetworkLossCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loss",
		Short: "loss network packet",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkLossAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "percentage of packets to drop (10 is 10%)")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "correlation is percentage (10 is 10%)")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")

	return cmd
}

func NewNetworkCorruptCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "corrupt",
		Short: "corrupt network packet",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkCorruptAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "percentage of packets to corrupt (10 is 10%)")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "correlation is percentage (10 is 10%)")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")

	return cmd
}

func NetworkDuplicateCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duplicate",
		Short: "duplicate network packet",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDuplicateAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "percentage of packets to duplicate (10 is 10%)")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "correlation is percentage (10 is 10%)")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010. "+
			"It can only be used in conjunction with -p tcp or -p udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")

	return cmd
}

func NetworkPartitionCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "partition",
		Short: "partition",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkPartitionAction
			options.CompleteDefaults()
			fx.New(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")
	cmd.Flags().StringVarP(&options.Direction, "direction", "", "both", "specifies the partition direction, values can be 'to', 'from' or 'both'. 'from' means packets coming from the 'IPAddress' or 'Hostname' and going to your server, 'to' means packets originating from your server and going to the 'IPAddress' or 'Hostname'.")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"only impact traffic using this IP protocol, supported: tcp, udp, icmp, all")
	cmd.Flags().StringVarP(&options.AcceptTCPFlags, "accept-tcp-flags", "", "", "only the packet which match the tcp flag can be accepted, others will be dropped. only set when the protocol is tcp.")

	return cmd
}

func NetworkDNSCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "attack DNS server or map specified host to specified IP",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDNSAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.DNSServer, "dns-server", "", "123.123.123.123",
		"update the DNS server in /etc/resolv.conf with this value")
	cmd.Flags().StringVarP(&options.DNSDomainName, "dns-domain-name", "d", "", "map this host to specified IP")
	cmd.Flags().StringVarP(&options.DNSIp, "dns-ip", "i", "", "map specified host to this IP address")

	return cmd
}

func NewNetworkBandwidthCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bandwidth",
		Short: "limit network bandwidth",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkBandwidthAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Rate, "rate", "r", "", "the speed knob, allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second")
	cmd.Flags().Uint32VarP(&options.Limit, "limit", "l", 0, "the number of bytes that can be queued waiting for tokens to become available")
	cmd.Flags().Uint32VarP(&options.Buffer, "buffer", "b", 0, "the maximum amount of bytes that tokens can be available for instantaneously")
	cmd.Flags().Uint64VarP(options.Peakrate, "peakrate", "", 0, "the maximum depletion rate of the bucket")
	cmd.Flags().Uint32VarP(options.Minburst, "minburst", "m", 0, "specifies the size of the peakrate bucket")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "only impact egress traffic to these IP addresses")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "only impact traffic to these hostnames")

	return cmd
}

func commonNetworkAttackFunc(options *core.NetworkCommand, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.NetworkAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack network successfully, uid: %s", uid))
}

func NewNetworkPortOccupiedCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port",
		Short: "attack network port",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.NetworkPortOccupiedAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Port, "port", "p", "", "this specified port is to occupied")
	return cmd
}

func NewNetworkFloodCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flood",
		Short: "generate a mount of network traffic by using iperf",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.NetworkFloodAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Rate, "rate", "r", "", "the speed of network traffic, allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "generate traffic to this IP address")
	cmd.Flags().StringVarP(&options.Port, "port", "p", "", "generate traffic to this port on the IP address")
	cmd.Flags().Int32VarP(&options.Parallel, "parallel", "", 1, "number of iperf parallel client threads to run")
	cmd.Flags().StringVarP(&options.Duration, "duration", "", "99999999", "number of seconds to run the iperf test")

	return cmd
}

func NewNICDownCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "down network interface card",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.NetworkNICDownAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "the network interface to impact")
	SetScheduleFlags(cmd, &options.SchedulerConfig)
	return cmd
}
