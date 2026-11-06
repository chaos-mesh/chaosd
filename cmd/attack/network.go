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
		Short: "Inject network faults",
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
		Short: "Delay outgoing network packets",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDelayAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Latency, "latency", "l", "",
		"Outgoing packet delay, e.g. 100ms or 1s (required)")
	cmd.Flags().StringVarP(&options.Jitter, "jitter", "j", "",
		"Variation in packet delay, e.g. 10ms")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "Delay correlation as a percentage, from 0 to 100")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"Filter outgoing traffic by destination ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"Filter outgoing traffic by source ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"Filter traffic by IP protocol: tcp, udp, icmp, or all")
	cmd.Flags().StringVarP(&options.AcceptTCPFlags, "accept-tcp-flags", "", "", "Currently unused for delay; requires --protocol tcp if set")

	return cmd
}

func NewNetworkLossCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loss",
		Short: "Drop outgoing network packets",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkLossAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "Percentage of packets to drop, from 0 to 100")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "Packet loss correlation as a percentage, from 0 to 100")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"Filter outgoing traffic by destination ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"Filter outgoing traffic by source ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"Filter traffic by IP protocol: tcp, udp, icmp, or all")

	return cmd
}

func NewNetworkCorruptCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "corrupt",
		Short: "Corrupt outgoing network packets",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkCorruptAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "Percentage of packets to corrupt, from 0 to 100")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "Packet corruption correlation as a percentage, from 0 to 100")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"Filter outgoing traffic by destination ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"Filter outgoing traffic by source ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"Filter traffic by IP protocol: tcp, udp, icmp, or all")

	return cmd
}

func NetworkDuplicateCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duplicate",
		Short: "Duplicate outgoing network packets",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDuplicateAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVar(&options.Percent, "percent", "1", "Percentage of packets to duplicate, from 0 to 100")
	cmd.Flags().StringVarP(&options.Correlation, "correlation", "c", "0", "Packet duplication correlation as a percentage, from 0 to 100")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	cmd.Flags().StringVarP(&options.EgressPort, "egress-port", "e", "",
		"Filter outgoing traffic by destination ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.SourcePort, "source-port", "s", "",
		"Filter outgoing traffic by source ports, e.g. 80,8001:8010; "+
			"requires --protocol tcp or udp")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"Filter traffic by IP protocol: tcp, udp, icmp, or all")

	return cmd
}

func NetworkPartitionCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "partition",
		Short: "Block network traffic to or from selected destinations",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkPartitionAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")
	cmd.Flags().StringVarP(&options.Direction, "direction", "", "both", "Choose which traffic to block: to selected destinations, from them, or both")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	cmd.Flags().StringVarP(&options.IPProtocol, "protocol", "p", "",
		"Filter traffic by IP protocol: tcp, udp, icmp, or all")
	cmd.Flags().StringVarP(&options.AcceptTCPFlags, "accept-tcp-flags", "", "", "Allow TCP packets matching an iptables --tcp-flags expression, e.g. \"SYN,ACK SYN\"; requires --protocol tcp")

	return cmd
}

func NetworkDNSCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Change DNS servers and optionally override hostname resolution",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkDNSAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.DNSServer, "dns-server", "", "123.123.123.123",
		"DNS server address in /etc/resolv.conf")
	cmd.Flags().StringVarP(&options.DNSDomainName, "dns-domain-name", "d", "", "Hostname to override in /etc/hosts; requires --dns-ip")
	cmd.Flags().StringVarP(&options.DNSIp, "dns-ip", "i", "", "IP address for --dns-domain-name in /etc/hosts")

	return cmd
}

func NewNetworkBandwidthCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bandwidth",
		Short: "Limit outgoing network bandwidth",

		Run: func(*cobra.Command, []string) {
			options.Action = core.NetworkBandwidthAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Rate, "rate", "r", "", "Bandwidth limit in bytes per second with a bps, kbps, mbps, gbps, or tbps suffix (required)")
	cmd.Flags().Uint32VarP(&options.Limit, "limit", "l", 0, "Maximum queued data in bytes (required, greater than 0)")
	cmd.Flags().Uint32VarP(&options.Buffer, "buffer", "b", 0, "Maximum burst size in bytes (required, greater than 0)")
	cmd.Flags().Uint64VarP(options.Peakrate, "peakrate", "", 0, "Optional peak transmission rate in bits per second; use with --minburst")
	cmd.Flags().Uint32VarP(options.Minburst, "minburst", "m", 0, "Peak-rate burst size in bytes; use with --peakrate")
	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Filter traffic by comma-separated IP addresses or CIDR ranges")
	cmd.Flags().StringVarP(&options.Hostname, "hostname", "H", "", "Filter traffic by comma-separated hostnames")

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
		Short: "Occupy a TCP port with an HTTP server",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.NetworkPortOccupiedAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Port, "port", "p", "", "TCP port to occupy (required)")
	return cmd
}

func NewNetworkFloodCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flood",
		Short: "Generate UDP traffic with an iperf client",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.NetworkFloodAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Rate, "rate", "r", "", "UDP traffic rate using iperf syntax, e.g. 10M for 10 Mbit/s (required)")
	cmd.Flags().StringVarP(&options.IPAddress, "ip", "i", "", "Destination iperf server IP address (required)")
	cmd.Flags().StringVarP(&options.Port, "port", "p", "", "Destination iperf server port (required)")
	cmd.Flags().Int32VarP(&options.Parallel, "parallel", "", 1, "Number of parallel iperf client streams")
	cmd.Flags().StringVarP(&options.Duration, "duration", "", "99999999", "How long to generate traffic, in seconds")

	return cmd
}

func NewNICDownCommand(dep fx.Option, options *core.NetworkCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Bring a network interface down",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.NetworkNICDownAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonNetworkAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Device, "device", "d", "", "Network interface to affect (required)")
	SetScheduleFlags(cmd, &options.SchedulerConfig)
	return cmd
}
