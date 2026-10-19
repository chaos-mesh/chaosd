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

func TestHTTPLongFlags(t *testing.T) {
	for _, flag := range []string{"--delay-time", "-d"} {
		t.Run(flag, func(t *testing.T) {
			o := core.NewHTTPAttackOption()
			cmd := NewHTTPDelayCommand(fx.Options(), o)
			if err := cmd.ParseFlags([]string{flag, "250ms"}); err != nil {
				t.Fatal(err)
			}
			if o.Delay != "250ms" {
				t.Fatalf("delay = %q", o.Delay)
			}
		})
	}
	for _, flag := range []string{"--file-path", "-p"} {
		t.Run(flag, func(t *testing.T) {
			o := core.NewHTTPAttackOption()
			cmd := NewHTTPConfigCommand(fx.Options(), o)
			if err := cmd.ParseFlags([]string{flag, "config.json"}); err != nil {
				t.Fatal(err)
			}
			if o.FilePath != "config.json" {
				t.Fatalf("file path = %q", o.FilePath)
			}
		})
	}
}
