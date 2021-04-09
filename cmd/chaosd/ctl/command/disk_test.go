// Copyright 2021 Chaos Mesh Authors.
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
	"os"
	"strconv"
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func TestServer_DiskFill(t *testing.T) {
	s := mustChaosdFromCmd(nil, &conf)
	tests := []struct {
		name    string
		fill    *core.DiskCommand
		wantErr bool
	}{
		{
			name: "0",
			fill: &core.DiskCommand{
				Action:          core.DiskFillAction,
				Size:            "1024",
				Path:            "temp",
				FillByFallocate: true,
			},
			wantErr: false,
		}, {
			name: "1",
			fill: &core.DiskCommand{
				Action:          core.DiskFillAction,
				Size:            "24",
				Path:            "temp",
				FillByFallocate: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Create(tt.fill.Path)
			if err != nil {
				t.Errorf("unexpected err %v when creating temp file", err)
			}
			if f != nil {
				_ = f.Close()
			}
			_, err = s.DiskFill(tt.fill)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiskFill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			stat, err := os.Stat(tt.fill.Path)
			if err != nil {
				t.Errorf("unexpected err %v when stat temp file", err)
			}

			if size, err := strconv.ParseUint(tt.fill.Size, 10, 0); err != nil {
				if uint64(stat.Size()) != size*1024*1024 {
					t.Errorf("DiskFill() size %v, expect %d", stat.Size(), size*1024*1024)
					return
				}
			}

			os.Remove(tt.fill.Path)
		})
	}
}

func TestServer_DiskPayload(t *testing.T) {
	s := mustChaosdFromCmd(nil, &conf)
	tests := []struct {
		name    string
		command *core.DiskCommand
		wantErr bool
	}{
		{
			name: "0",
			command: &core.DiskCommand{
				Action: core.DiskWritePayloadAction,
				Size:   "24",
				Path:   "temp",
			},
			wantErr: false,
		}, {
			name: "1",
			command: &core.DiskCommand{
				Action: core.DiskReadPayloadAction,
				Size:   "24",
				Path:   "temp",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Create(tt.command.Path)
			if err != nil {
				t.Errorf("unexpected err %v when creating temp file", err)
			}
			if f != nil {
				_ = f.Close()
			}

			_, err = s.DiskFill(&core.DiskCommand{
				Action:          core.DiskFillAction,
				Size:            tt.command.Size,
				Path:            "temp",
				FillByFallocate: true,
			})
			_, err = s.DiskPayload(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiskPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			os.Remove(tt.command.Path)
		})
	}
}
