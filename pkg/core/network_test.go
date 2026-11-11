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
	"encoding/json"
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
					Action:     NetworkPartitionAction,
					Direction:  "to",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					Action:     NetworkPartitionAction,
					Direction:  "from",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "INPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					Action:     NetworkPartitionAction,
					Direction:  "both",
					IPProtocol: "tcp",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
					{
						Name:      "INPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
			{
				cmd: &NetworkCommand{
					Action:         NetworkPartitionAction,
					Direction:      "both",
					IPProtocol:     "tcp",
					AcceptTCPFlags: "SYN,ACK SYN,ACK",
				},
				chains: []*pb.Chain{
					{
						Name:      "OUTPUT/3c552_e0172bc4fd04_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						TcpFlags:  "SYN,ACK SYN,ACK",
						Target:    "ACCEPT",
					},
					{
						Name:      "OUTPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_OUTPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
					{
						Name:      "INPUT/3c552_e0172bc4fd04_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						TcpFlags:  "SYN,ACK SYN,ACK",
						Target:    "ACCEPT",
					},
					{
						Name:      "INPUT/3c552_e0172bc4fd046_",
						Ipsets:    []string{"test"},
						Direction: pb.Chain_INPUT,
						Protocol:  "tcp",
						Target:    "DROP",
					},
				},
			},
		}
		for _, tc := range testCases {
			chains, err := tc.cmd.AdditionalChain("test", "eth0", "3c5528e1-4c32-4f80-983c-913ad7e860e2")
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

func TestNetworkDurationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, action, duration, schedule string
		invalid                          bool
	}{
		{"flood seconds", NetworkFloodAction, "10", "", false},
		{"flood default", NetworkFloodAction, "99999999", "", false},
		{"down seconds", NetworkNICDownAction, "10", "", false},
		{"down indefinite", NetworkNICDownAction, "-1", "", false},
		{"scheduled flood seconds", NetworkFloodAction, "10", "* * * * *", true},
		{"scheduled down indefinite", NetworkNICDownAction, "-1", "* * * * *", true},
		{"scheduled invalid duration", NetworkFloodAction, "invalid", "* * * * *", true},
		{"scheduled empty duration", NetworkNICDownAction, "", "* * * * *", true},
		{"scheduled Go duration", NetworkNICDownAction, "10s", "* * * * *", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NetworkCommand{
				Action:   tc.action,
				Duration: tc.duration, Schedule: tc.schedule,
				Device:        "eth0",
				IPAddress:     "127.0.0.1",
				Port:          "5001",
				BandwidthSpec: &BandwidthSpec{Rate: "1mbps"},
			}
			if err := cmd.Validate(); (err != nil) != tc.invalid {
				t.Errorf("duration=%q schedule=%q: %v", tc.duration, tc.schedule, err)
			}
		})
	}
}

func TestNetworkDelayReorder(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		for _, tc := range []struct {
			name, latency, reorder, correlation string
			gap                                 int
			invalid                             bool
		}{
			{"without reorder", "10ms", "", "", 0, false},
			{"reorder", "10ms", "25", "", 0, false},
			{"reorder with correlation and gap", "10ms", "25", "50", 5, false},
			{"invalid reorder", "10ms", "101", "", 0, true},
			{"invalid correlation", "10ms", "25", "abc", 0, true},
			{"negative gap", "10ms", "25", "", -1, true},
			{"zero latency", "0s", "25", "", 0, true},
			{"correlation without reorder", "10ms", "", "50", 0, true},
			{"gap without reorder", "10ms", "", "", 5, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				cmd := &NetworkCommand{
					Action:             NetworkDelayAction,
					Latency:            tc.latency,
					Device:             "eth0",
					Reorder:            tc.reorder,
					ReorderCorrelation: tc.correlation,
					ReorderGap:         tc.gap,
				}
				cmd.CompleteDefaults()
				if err := cmd.Validate(); (err != nil) != tc.invalid {
					t.Fatalf("invalid=%v, got error %v", tc.invalid, err)
				}
			})
		}
	})

	cmd := &NetworkCommand{
		Action:     NetworkDelayAction,
		Latency:    "10ms",
		Device:     "eth0",
		Reorder:    "25",
		ReorderGap: 5,
	}
	cmd.CompleteDefaults()

	tc, err := cmd.ToTC("")
	if err != nil {
		t.Fatal(err)
	}
	checkReorder := func(netem *pb.Netem) {
		t.Helper()
		if netem.Time != "10ms" || netem.Reorder != 25 || netem.ReorderCorr != 0 || netem.Gap != 5 {
			t.Fatalf("netem=%v, want 10ms delay with reorder 25 and gap 5", netem)
		}
	}
	checkReorder(tc.Netem)

	// Rules of other experiments on the same device are restored from their
	// stored parameters, so they must keep the reorder settings.
	data, err := json.Marshal(&TcParameter{Device: cmd.Device, Delay: cmd.ToDelaySpec()})
	if err != nil {
		t.Fatal(err)
	}
	restored, err := (&TCRule{Type: pb.Tc_NETEM.String(), TC: string(data)}).ToTC()
	if err != nil {
		t.Fatal(err)
	}
	checkReorder(restored.Netem)
}
