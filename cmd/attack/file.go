// Copyright 2022 Chaos Mesh Authors.
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

func NewFileAttackCommand(uid *string) *cobra.Command {
	options := core.NewFileCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.FileCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "file <subcommand>",
		Short: "Inject file and directory faults",
	}

	cmd.AddCommand(
		NewFileCreateCommand(dep, options),
		NewFileModifyPrivilegeCommand(dep, options),
		NewFileDeleteCommand(dep, options),
		NewFileRenameCommand(dep, options),
		NewFileAppendCommand(dep, options),
		NewFileReplaceCommand(dep, options),
	)

	return cmd
}

func NewFileCreateCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a file or directory",

		Run: func(*cobra.Command, []string) {
			options.Action = core.FileCreateAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "file-name", "f", "", "File path to create (truncates an existing file)")
	cmd.Flags().StringVarP(&options.DirName, "dir-name", "d", "", "Directory path to create (takes precedence over --file-name)")

	return cmd
}

func NewFileModifyPrivilegeCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Change file permissions",
		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.FileModifyPrivilegeAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "file-name", "f", "", "File path whose permissions to change")
	cmd.Flags().Uint32VarP(&options.Privilege, "privilege", "p", 0, "Permission digits without a leading zero, such as 644 or 755")

	return cmd
}

func NewFileDeleteCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Make a file or directory unavailable by renaming it",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.FileDeleteAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "file-name", "f", "", "File path to remove from its original location")
	cmd.Flags().StringVarP(&options.DirName, "dir-name", "d", "", "Directory path to remove when --file-name is omitted")

	return cmd
}

func NewFileRenameCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename a file or directory",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.FileRenameAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.SourceFile, "source-file", "s", "", "Current path of the file or directory")
	cmd.Flags().StringVarP(&options.DestFile, "dest-file", "d", "", "New path of the file or directory")

	return cmd
}

func NewFileAppendCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append",
		Short: "Append data to a file",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.FileAppendAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "file-name", "f", "", "Existing file path to append to")
	cmd.Flags().StringVarP(&options.Data, "data", "d", "", "Data to append, followed by a newline")
	cmd.Flags().IntVarP(&options.Count, "count", "c", 1, "Number of times to append the data")

	return cmd
}

func NewFileReplaceCommand(dep fx.Option, options *core.FileCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Replace matching text in a file",

		Run: func(_ *cobra.Command, _ []string) {
			options.Action = core.FileReplaceAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(commonFileAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.FileName, "file-name", "f", "", "File path to modify")
	cmd.Flags().StringVarP(&options.OriginStr, "origin-string", "o", "", "sed basic regular expression to replace")
	cmd.Flags().StringVarP(&options.DestStr, "dest-string", "d", "", "Replacement text using sed replacement syntax")
	cmd.Flags().IntVarP(&options.Line, "line", "l", 0, "Line number to modify, starting at 1 (0 selects all lines)")

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
