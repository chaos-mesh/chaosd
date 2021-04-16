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
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

func NewDiskAttackCommand() *cobra.Command {
	options := core.NewDiskOption()
	dep := fx.Options(
		Module,
		fx.Provide(func() *core.DiskOption {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "disk <subcommand>",
		Short: "disk attack related command",
	}
	cmd.AddCommand(
		NewDiskPayloadCommand(dep, options),
		NewDiskFillCommand(dep, options),
	)
	return cmd
}

func NewDiskPayloadCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-payload <subcommand>",
		Short: "add disk payload",
	}

	cmd.AddCommand(
		NewDiskWritePayloadCommand(dep, options),
		NewDiskReadPayloadCommand(dep, options),
	)

	return cmd
}

func NewDiskWritePayloadCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write",
		Short: "write payload",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskWritePayloadAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will fill in the file path with unit.")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "/dev/null",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, payload will write into /dev/null")
	cmd.Flags().StringVarP(&options.Unit, "unit", "u", "M",
		"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES, default : M")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "pn", 1,
		"'process-num' specifies the number of process work on writing , default 1, only 1-255 is valid value")
	commonFlags(cmd, &options.CommonAttackConfig)
	return cmd
}

func NewDiskReadPayloadCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "read payload",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskReadPayloadAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will read from the file path with unit.")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to read data.\n"+
			"If path not provided, payload will raise an error")
	cmd.Flags().StringVarP(&options.Unit, "unit", "u", "M",
		"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES, default : M")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "pn", 1,
		"'process-num' specifies the number of process work on reading , default 1, only 1-255 is valid value")
	commonFlags(cmd, &options.CommonAttackConfig)
	return cmd
}

func NewDiskFillCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "fill disk",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskFillAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will fill in the file path with unit.")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, a temp file will be generated and deleted immediately after data filled in or allocated")
	cmd.Flags().StringVarP(&options.Percent, "percent", "c", "",
		"'percent' how many percent data of disk will fill in the file path")
	cmd.Flags().StringVarP(&options.Unit, "unit", "u", "M",
		"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES")
	cmd.Flags().BoolVarP(&options.FillByFallocate, "fallocate", "f", true, "fill disk by fallocate instead of dd")
	cmd.Flags().BoolVarP(&options.FillDestroyFile, "destroy", "d", false, "destroy file after filled in or allocated")
	commonFlags(cmd, &options.CommonAttackConfig)
	return cmd
}

func processDiskAttack(options *core.DiskOption, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}
	uid, err := chaos.ExecuteAttack(chaosd.DiskAttack, options)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	if options.String() == core.DiskWritePayloadAction {
		NormalExit(fmt.Sprintf("Write file %s successfully, uid: %s", options.Path, uid))
	} else if options.String() == core.DiskReadPayloadAction {
		NormalExit(fmt.Sprintf("Read file %s successfully, uid: %s", options.Path, uid))
	} else {
		NormalExit(fmt.Sprintf("Fill file %s successfully, uid: %s", options.Path, uid))
	}
}
