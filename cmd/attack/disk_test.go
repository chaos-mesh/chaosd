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

package attack

import (
	"os"
	"strconv"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

type diskTest struct {
	name    string
	option  *core.DiskOption
	wantErr bool
}

func TestServer_DiskFill(t *testing.T) {
	fxtest.New(
		t,
		server.Module,
		fx.Provide(func() []diskTest {
			return []diskTest{
				{
					name: "0",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskFillAction,
							Kind:   core.DiskAttack,
						},
						Size:            "1024",
						Path:            "temp",
						Unit:            "M",
						FillByFallocate: true,
					},
					wantErr: false,
				}, {
					name: "1",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskFillAction,
							Kind:   core.DiskAttack,
						},
						Size:            "24",
						Path:            "temp",
						Unit:            "MB",
						FillByFallocate: false,
					},
					wantErr: false,
				},
			}
		}),
		fx.Invoke(func(s *chaosd.Server, tests []diskTest) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					f, err := os.Create(tt.option.Path)
					if err != nil {
						t.Errorf("unexpected err %v when creating temp file", err)
						return
					}
					if f != nil {
						_ = f.Close()
					}
					_, err = s.ExecuteAttack(chaosd.DiskAttack, tt.option)
					if (err != nil) != tt.wantErr {
						t.Errorf("DiskFill() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					stat, err := os.Stat(tt.option.Path)
					if err != nil {
						t.Errorf("unexpected err %v when stat temp file", err)
						return
					}

					size, _ := strconv.ParseUint(tt.option.Size, 10, 0)
					usize := uint64(0)
					if tt.option.Unit == "M" {
						usize = 1024 * 1024
					} else if tt.option.Unit == "MB" {
						usize = 1000 * 1000
					}
					if stat.Size() != int64(size*usize) {
						t.Errorf("DiskFill() size %v, expect %d", stat.Size(), size*usize)
						return
					}
					os.Remove(tt.option.Path)
				})
			}
		}),
	)
}

func TestServer_DiskPayload(t *testing.T) {
	fxtest.New(
		t,
		server.Module,
		fx.Provide(func() []diskTest {
			return []diskTest{
				{
					name: "0",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskWritePayloadAction,
							Kind:   core.DiskAttack,
						},
						Size: "24",
						Path: "temp",
						Unit: "M",
					},
					wantErr: false,
				}, {
					name: "1",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskReadPayloadAction,
							Kind:   core.DiskAttack,
						},
						Size: "24",
						Path: "temp",
						Unit: "M",
					},
					wantErr: false,
				},
			}
		}),
		fx.Invoke(func(s *chaosd.Server, tests []diskTest) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					f, err := os.Create(tt.option.Path)
					if err != nil {
						t.Errorf("unexpected err %v when creating temp file", err)
						return
					}
					if f != nil {
						_ = f.Close()
					}

					_, err = s.ExecuteAttack(chaosd.DiskAttack, &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskFillAction,
							Kind:   core.DiskAttack,
						},
						Size:            tt.option.Size,
						Path:            "temp",
						FillByFallocate: true,
					})
					_, err = s.ExecuteAttack(chaosd.DiskAttack, tt.option)
					if (err != nil) != tt.wantErr {
						t.Errorf("DiskPayload() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					os.Remove(tt.option.Path)
				})
			}
		}),
	)
}
