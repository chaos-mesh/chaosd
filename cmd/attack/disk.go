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

func NewDiskAttackCommand(uid *string) *cobra.Command {
	options := core.NewDiskOption()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.DiskOption {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "disk <subcommand>",
		Short: "Inject disk load or fill disk space",
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
		Short: "Generate disk I/O load",
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
		Short: "Generate disk write load",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskWritePayloadAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"Total data to write across workers, such as 1024c, 1MiB, or 1MB (bare numbers use MiB)")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"New file path or existing directory for output (defaults to the current directory)")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "n", 1,
		"Number of write workers, from 1 to 255")
	SetScheduleFlags(cmd, &options.SchedulerConfig)
	return cmd
}

func NewDiskReadPayloadCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Generate disk read load",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskReadPayloadAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"Total data to read across workers, such as 1024c, 1MiB, or 1MB (bare numbers use MiB)")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"File or device to read (defaults to the device mounted at /)")
	cmd.Flags().Uint8VarP(&options.PayloadProcessNum, "process-num", "n", 1,
		"Number of read workers, from 1 to 255")
	SetScheduleFlags(cmd, &options.SchedulerConfig)
	return cmd
}

func NewDiskFillCommand(dep fx.Option, options *core.DiskOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "Fill disk space",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskFillAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"Data size to allocate or write, such as 1024c, 1MiB, or 1MB (bare numbers use MiB)")
	cmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"New file path or existing directory to fill (defaults to the current directory)")
	cmd.Flags().StringVarP(&options.Percent, "percent", "c", "",
		"Percentage of total filesystem capacity to fill when --size is omitted")
	cmd.Flags().BoolVarP(&options.FillByFallocate, "fallocate", "f", true, "Allocate disk space with fallocate instead of writing zeros with dd")
	return cmd
}

func processDiskAttack(options *core.DiskOption, chaos *chaosd.Server) {
	attackConfig, err := options.PreProcess()
	if err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.DiskAttack, attackConfig, core.CommandMode)
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
