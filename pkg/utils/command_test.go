// Copyright 2023 Chaos Mesh Authors.
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

package utils

import (
	"fmt"
	"testing"
)

func TestCommand_Unmarshal(t *testing.T) {
	type dd struct {
		If    string `dd:"if"`
		Of    string `dd:"oflag"`
		Iflag string `dd:"iflag"`
	}
	dc := Command{Name: "dd"}
	tests := []struct {
		name string
		d    dd
	}{
		{
			name: "0",
			d: dd{
				"/dev/zero",
				"i,2,3",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := dc.Unmarshal(tt.d)
			fmt.Println(cmd.String())
		})
	}
}
