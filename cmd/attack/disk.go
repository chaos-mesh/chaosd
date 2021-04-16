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
	"github.com/chaos-mesh/chaosd/cmd/server"

	"github.com/chaos-mesh/chaosd/pkg/utils"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

func NewDiskAttackCommand() *cobra.Command {
	options := core.NewDiskCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.DiskCommand {
			return options
		}),
	)

	diskCmd := &cobra.Command{
		Use:   "disk <subcommand>",
		Short: "disk attack related command",
	}
	diskCmd.AddCommand(
		NewDiskPayloadCommand(dep, options),
		NewDiskFillCommand(dep, options),
	)
	return diskCmd
}

func NewDiskPayloadCommand(dep fx.Option, options *core.DiskCommand) *cobra.Command {
	addPayloadCmd := &cobra.Command{
		Use:   "add-payload <subcommand>",
		Short: "add disk payload",
	}

	addPayloadCmd.AddCommand(
		NewDiskWritePayloadCommand(dep, options),
		NewDiskReadPayloadCommand(dep, options),
	)

	return addPayloadCmd
}

func NewDiskWritePayloadCommand(dep fx.Option, options *core.DiskCommand) *cobra.Command {
	writeCmd := &cobra.Command{
		Use:   "write",
		Short: "write payload",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskWritePayloadAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	writeCmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will fill in the file path with unit MB.")
	writeCmd.Flags().StringVarP(&options.Path, "path", "p", "/dev/null",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, payload will write into /dev/null")
	return writeCmd
}

func NewDiskReadPayloadCommand(dep fx.Option, options *core.DiskCommand) *cobra.Command {
	readCmd := &cobra.Command{
		Use:   "read",
		Short: "read payload",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskReadPayloadAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	readCmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will read from the file path with unit MB.")
	readCmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to read data.\n"+
			"If path not provided, payload will raise an error")
	return readCmd
}

func NewDiskFillCommand(dep fx.Option, options *core.DiskCommand) *cobra.Command {
	fillCmd := &cobra.Command{
		Use:   "fill",
		Short: "fill disk",
		Run: func(*cobra.Command, []string) {
			options.Action = core.DiskFillAction
			fx.New(dep, fx.Invoke(processDiskAttack)).Run()
		},
	}

	fillCmd.Flags().StringVarP(&options.Size, "size", "s", "",
		"'size' specifies how many data will fill in the file path with unit MB.")
	fillCmd.Flags().StringVarP(&options.Path, "path", "p", "",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, a temp file will be generated and deleted immediately after data filled in or allocated")
	fillCmd.Flags().StringVarP(&options.Percent, "percent", "c", "",
		"'percent' how many percent data of disk will fill in the file path")
	fillCmd.Flags().BoolVarP(&options.FillByFallocate, "fallocate", "f", true, "fill disk by fallocate instead of dd")
	return fillCmd
}

func processDiskAttack(options *core.DiskCommand, chaos *chaosd.Server) {
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
