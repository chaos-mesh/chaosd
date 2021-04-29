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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
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
						Size:              "1024M",
						Path:              "temp",
						FillByFallocate:   true,
						PayloadProcessNum: 1,
					},
					wantErr: false,
				}, {
					name: "1",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskFillAction,
							Kind:   core.DiskAttack,
						},
						Size:              "24MB",
						Path:              "temp",
						FillByFallocate:   false,
						PayloadProcessNum: 1,
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

					size, _ := utils.ParseUnit(tt.option.Size)
					if stat.Size() != int64(size) {
						t.Errorf("DiskFill() size %v, expect %d", stat.Size(), size)
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
						Size:              "24M",
						Path:              "temp",
						PayloadProcessNum: 1,
					},
					wantErr: false,
				}, {
					name: "1",
					option: &core.DiskOption{
						CommonAttackConfig: core.CommonAttackConfig{
							Action: core.DiskReadPayloadAction,
							Kind:   core.DiskAttack,
						},
						Size:              "24M",
						Path:              "temp",
						PayloadProcessNum: 1,
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

type writeArgs struct {
	Size              string
	Path              string
	PayloadProcessNum uint8
}

func writeArgsToDiskOption(args writeArgs) core.DiskOption {
	return core.DiskOption{
		CommonAttackConfig: core.CommonAttackConfig{
			SchedulerConfig: core.SchedulerConfig{},
			Action:          core.DiskWritePayloadAction,
			Kind:            "",
		},
		Size:              args.Size,
		Path:              args.Path,
		Percent:           "",
		FillByFallocate:   false,
		FillDestroyFile:   false,
		PayloadProcessNum: args.PayloadProcessNum,
	}
}

func writeArgsAttack(args writeArgs) error {
	opt := writeArgsToDiskOption(args)
	return chaosd.DiskAttack.Attack(&opt, chaosd.Environment{})
}

func TestNewDiskWritePayloadCommand(t *testing.T) {
	var opt core.DiskOption
	var err error
	opt = writeArgsToDiskOption(writeArgs{
		Size:              "",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "one of percent and size must not be empty, DiskOption : write-payload")

	opt = writeArgsToDiskOption(writeArgs{
		Size:              "1Ms",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "unknown units of size : 1Ms, DiskOption : write-payload")

	opt = writeArgsToDiskOption(writeArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "unsupport process num : 0, DiskOption : write-payload")

	opt = writeArgsToDiskOption(writeArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 1,
	})
	err = opt.Validate()
	assert.NoError(t, err)

	assert.NoError(t, writeArgsAttack(writeArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 1,
	}))

	assert.NoError(t, writeArgsAttack(writeArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 255,
	}))

	assert.NoError(t, writeArgsAttack(writeArgs{
		Size:              "1",
		Path:              "",
		PayloadProcessNum: 2,
	}))

	assert.Error(t, writeArgsAttack(writeArgs{
		Size:              "1",
		Path:              "&^%$#@#$%^&*(",
		PayloadProcessNum: 5,
	}))
}

type readArgs struct {
	Size              string
	Path              string
	PayloadProcessNum uint8
}

func readArgsToDiskOption(args readArgs) core.DiskOption {
	return core.DiskOption{
		CommonAttackConfig: core.CommonAttackConfig{
			SchedulerConfig: core.SchedulerConfig{},
			Action:          core.DiskReadPayloadAction,
			Kind:            "",
		},
		Size:              args.Size,
		Path:              args.Path,
		Percent:           "",
		FillByFallocate:   false,
		FillDestroyFile:   false,
		PayloadProcessNum: args.PayloadProcessNum,
	}
}

func readArgsAttack(args readArgs) error {
	opt := readArgsToDiskOption(args)
	return chaosd.DiskAttack.Attack(&opt, chaosd.Environment{})
}

func TestNewDiskReadPayloadCommand(t *testing.T) {
	var opt core.DiskOption
	var err error
	opt = readArgsToDiskOption(readArgs{
		Size:              "",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "one of percent and size must not be empty, DiskOption : read-payload")

	opt = readArgsToDiskOption(readArgs{
		Size:              "1Ms",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "unknown units of size : 1Ms, DiskOption : read-payload")

	opt = readArgsToDiskOption(readArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 0,
	})
	err = opt.Validate()
	assert.EqualError(t, err, "unsupport process num : 0, DiskOption : read-payload")

	opt = readArgsToDiskOption(readArgs{
		Size:              "0",
		Path:              "",
		PayloadProcessNum: 1,
	})
	err = opt.Validate()
	assert.NoError(t, err)

	assert.NoError(t, readArgsAttack(readArgs{
		Size:              "0",
		Path:              "/dev/zero",
		PayloadProcessNum: 1,
	}))

	assert.NoError(t, readArgsAttack(readArgs{
		Size:              "1",
		Path:              "/dev/zero",
		PayloadProcessNum: 2,
	}))
}
