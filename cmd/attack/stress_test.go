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

package attack

import (
	"testing"

	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func TestStressMemFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		workers int
	}{
		{name: "long", args: []string{"--workers", "2", "--size", "100M"}, workers: 2},
		{name: "short", args: []string{"-w", "2", "-s", "100M"}, workers: 2},
		{name: "default", args: []string{"--size", "100M"}, workers: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := core.NewStressCommand()
			cmd := NewStressMemCommand(fx.Options(), o)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			if o.Workers != tc.workers {
				t.Fatalf("workers = %d, want %d", o.Workers, tc.workers)
			}
			if o.Size != "100M" {
				t.Fatalf("size = %q, want 100M", o.Size)
			}
		})
	}
}

func TestStressAttackMemFlags(t *testing.T) {
	uid := ""
	cmd := NewStressAttackCommand(&uid)
	mem, args, err := cmd.Find([]string{"mem", "--workers", "2", "--size", "100M"})
	if err != nil {
		t.Fatal(err)
	}
	if err := mem.ParseFlags(args); err != nil {
		t.Fatal(err)
	}
	workers, err := mem.Flags().GetInt("workers")
	if err != nil {
		t.Fatal(err)
	}
	if workers != 2 {
		t.Fatalf("workers = %d, want 2", workers)
	}
	size, err := mem.Flags().GetString("size")
	if err != nil {
		t.Fatal(err)
	}
	if size != "100M" {
		t.Fatalf("size = %q, want 100M", size)
	}
}
