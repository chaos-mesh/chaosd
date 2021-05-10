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
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func NewDiskAttackCommand() *cobra.Command {
	options := core.NewDiskOption()
	dep := fx.Options(
		server.Module,
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
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many units of data will write into the file path."+
			"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES"+
			"example : 1M | 512kB")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to fill data in."+
			"If path not provided, payload will write into a temp file, temp file will be deleted after writing")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "n", 1,
		"'process-num' specifies the number of process work on writing , default 1, only 1-255 is valid value")
	return cmd
}

func NewDiskReadPayloadCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "read payload",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskReadPayloadAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many units of data will read from the file path."+
			"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES"+
			"example : 1M | 512kB")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to read data."+
			"If path not provided, payload will read from disk mount on \"/\"")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "n", 1,
		"'process-num' specifies the number of process work on reading , default 1, only 1-255 is valid value")
	return cmd
}

func NewDiskFillCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "fill disk",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskFillAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack), fx.NopLogger).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many units of data will fill in the file path."+
			"'unit' specifies the unit of data, support c=1, w=2, b=512, kB=1000, K=1024, MB=1000*1000,"+
			"M=1024*1024, , GB=1000*1000*1000, G=1024*1024*1024 BYTES"+
			"example : 1M | 512kB")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to fill data in."+
			"If path not provided, a temp file will be generated and deleted immediately after data filled in or allocated")
	cmd.Flags().StringVarP(&options.Percent, "percent", "c", "",
		"'percent' how many percent data of disk will fill in the file path")
	cmd.Flags().BoolVarP(&options.FillByFallocate, "fallocate", "f", true, "fill disk by fallocate instead of dd")
	cmd.Flags().BoolVarP(&options.DestroyFile, "destroy", "d", false, "destroy file after filled in or allocated")
	return cmd
}

func processDiskAttack(options *core.DiskOption, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}
	uid, err := chaos.ExecuteAttack(chaosd.DiskAttack, options)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	if options.String() == core.DiskWritePayloadAction {
		utils.NormalExit(fmt.Sprintf("Write file %s successfully, uid: %s", options.Path, uid))
	} else if options.String() == core.DiskReadPayloadAction {
		utils.NormalExit(fmt.Sprintf("Read file %s successfully, uid: %s", options.Path, uid))
	} else {
		utils.NormalExit(fmt.Sprintf("Fill file %s successfully, uid: %s", options.Path, uid))
	}
}
