// Copyright 2022 Chaos Mesh Authors.
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
	"testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestPatitionChain(t *testing.T) {
	t.Run("PattitionOnDirection", func(t *testing.T) {
		testCases := []struct {
			cmd    *NetworkCommand
			chains []*pb.Chain
		}{
			{
				cmd: &NetworkCommand{
					CommonAttackConfig: CommonAttackConfig{
						Action: NetworkPartitionAction,
					},
					Direction:  "to",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					CommonAttackConfig: CommonAttackConfig{
						Action: NetworkPartitionAction,
					},
					Direction:  "from",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "INPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					CommonAttackConfig: CommonAttackConfig{
						Action: NetworkPartitionAction,
					},
					Direction:  "both",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
					{
						Name:      "INPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					CommonAttackConfig: CommonAttackConfig{
						Action: NetworkPartitionAction,
					},
					Direction:      "both",
					IPProtocol:     "tcp",
					AcceptTCPFlags: "SYN,ACK SYN,ACK",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/0",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						TcpFlags:  "SYN,ACK SYN,ACK",
						Target:    "ACCEPT",
					},
					{
						Name:      "OUTPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
					{
						Name:      "INPUT/0",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						TcpFlags:  "SYN,ACK SYN,ACK",
						Target:    "ACCEPT",
					},
					{
						Name:      "INPUT/1",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
		}
		for _, tc := range testCases {
			chains, err := tc.cmd.PartitionChain("test")
			if err != nil {
				t.Errorf("failed to partition chain: %v", err)
			}
			if len(chains) != len(tc.chains) {
				t.Errorf("invalid chains. expected: %v, actual: %v", tc.chains, chains)
			}
			for i, chain := range chains {
				if chain.Name != tc.chains[i].Name {
					t.Errorf("invalid chain name. expected: %v, actual: %v", tc.chains[i].Name, chain.Name)
				}
				if chain.Ipsets[0] != "test" {
					t.Errorf("invalid ipsets. expected: %v, actual: %v", tc.chains[i].Ipsets, chain.Ipsets)
				}
				if chain.Direction != tc.chains[i].Direction {
					t.Errorf("invalid direction. expected: %v, actual: %v", tc.chains[i].Direction, chain.Direction)
				}
				if chain.Target != tc.chains[i].Target {
					t.Errorf("invalid target. expected: %v, actual: %v", tc.chains[i].Target, chain.Target)
				}
			}
		}
	})
}
