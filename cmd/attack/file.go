// Copyright 2020 Chaos Mesh Authors.
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

func NewFileAttackCommand() *cobra.Command {
	options := core.NewFileCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.FileCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "file <subcommand>",
		Short: "File attack related commands",
	}

	cmd.AddCommand(
		NewFileCreateCommand(dep, options),
		NewFileModifyPrivilegeCommand(dep, options),
		NewFileDeleteCommand(dep, options),
		NewFileRenameCommand(dep, options),
		NewFileAppendCommand(dep, options),
	)

	return cmd
}

func NewFileCreateCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create file",

		Run: func(*cobra.Command, []string) {
			options.Action = core.FileCreateAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "filename", "f", "", "create file based on filename")
	cmd.Flags().StringVarP(&options.DirName, "dirname", "d", "", "create directory based on dirname")
	cmd.Flags().StringVarP(&options.DestDir, "destdir", "", "", "create a file or directory to the specified destdir")
	// owner TODO

	return cmd
}

func NewFileModifyPrivilegeCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "modify file privilege",
		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.FileModifyPrivilegeAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "filename", "f", "", "file to be change privilege")
	cmd.Flags().Uint32VarP(&options.Privilege, "privilege", "p", 0, "privilege to be update")

	return cmd
}

func NewFileDeleteCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete file",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.FileDeleteAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "filename", "f", "", "delete file based on filename")
	cmd.Flags().StringVarP(&options.DirName, "dirname", "d", "", "delete directory based on dirname")
	cmd.Flags().StringVarP(&options.DestDir, "destdir", "", "", "delete a file or directory to the specified destdir")

	return cmd
}

func NewFileRenameCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "rename file",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.FileRenameAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.SourceFile, "source-file", "s", "", "the source file/dir of rename")
	cmd.Flags().StringVarP(&options.DstFile, "dst-file", "d", "", "the destination file/dir of rename")

	return cmd
}

func NewFileAppendCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use: "append",
		Short: "append file",

		Run: func(cmd *cobra.Command, args []string) {
			options.Action = core.FileAppendAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "filename", "f", "", "append data to the file")
	cmd.Flags().StringVarP(&options.Data, "data", "d", "", "append data")
	cmd.Flags().IntVarP(&options.Count, "count", "c", 1, "append count with default value is 1")
	cmd.Flags().IntVarP(&options.LineNo, "line", "l", 0, "the start line to append with default value is 1")

	return cmd
}

func commonFileAttackFunc(options *core.FileCommand, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.FileAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack file successfully, uid: %s", uid))
}
