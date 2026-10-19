// Copyright 2026 Chaos Mesh Authors.
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
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestNetworkDelayDurationValues(t *testing.T) {
	for _, tc := range []struct{ latency, jitter string }{
		{"1500ms", "250us"}, {"2h", "500ns"}, {"0s", "0s"},
	} {
		t.Run(tc.latency, func(t *testing.T) {
			command := &NetworkCommand{Latency: tc.latency, Jitter: tc.jitter, Correlation: "25"}
			spec := &DelaySpec{Latency: tc.latency, Jitter: tc.jitter, Correlation: "25"}
			for name, convert := range map[string]func() (*pb.Netem, error){"command": command.ToDelayNetem, "spec": spec.ToNetem} {
				t.Run(name, func(t *testing.T) {
					netem, err := convert()
					if err != nil {
						t.Fatal(err)
					}
					for field, pair := range map[string][2]string{"delay": {netem.Time, tc.latency}, "jitter": {netem.Jitter, tc.jitter}} {
						got, err := time.ParseDuration(pair[0])
						want, _ := time.ParseDuration(pair[1])
						if err != nil || got != want {
							t.Fatalf("%s=%q, want duration %q", field, pair[0], pair[1])
						}
					}
					if netem.DelayCorr != 25 {
						t.Fatalf("correlation=%v, want 25", netem.DelayCorr)
					}
				})
			}
		})
	}
}

func TestNetworkMergedDurationValues(t *testing.T) {
	netem, err := toNetem(&TcParameter{
		Delay: &DelaySpec{Latency: "2s", Jitter: "500us", Correlation: "25"},
		Loss:  &LossSpec{Loss: "10", Correlation: "50"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if netem.Time != "2s" || netem.Jitter != "500µs" || netem.Loss != 10 || netem.LossCorr != 50 || netem.DelayCorr != 25 {
		t.Fatalf("merged delay/loss=%v", netem)
	}
	for _, reversed := range []bool{false, true} {
		a := &pb.Netem{Time: "900ms", Jitter: "900us"}
		b := &pb.Netem{Time: "1s", Jitter: "1ms"}
		if reversed {
			a, b = b, a
		}
		merged := mergeNetem(a, b)
		if merged.Time != "1s" || merged.Jitter != "1ms" {
			t.Fatalf("reversed=%v: merged durations=%v", reversed, merged)
		}
	}
}

func TestBandwidthRateStringPreservesNumericValue(t *testing.T) {
	for _, tc := range []struct{ rate, want string }{
		{"7bps", "7"}, {"2kbps", "2048"}, {"3mbps", "3145728"},
		{"1gbps", "1073741824"}, {"1tbps", "1099511627776"},
	} {
		t.Run(tc.rate, func(t *testing.T) {
			tbf, err := (&BandwidthSpec{Rate: tc.rate, Limit: 10000, Buffer: 1000}).ToTbf()
			if err != nil {
				t.Fatal(err)
			}
			if tbf.Rate != tc.want || tbf.Limit != 10000 || tbf.Buffer != 1000 {
				t.Fatalf("tbf=%v, want rate=%s limit=10000 buffer=1000", tbf, tc.want)
			}
		})
	}
}
