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

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var dFlag core.DiskCommand

func NewDiskAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk <subcommand>",
		Short: "disk attack related command",
	}
	cmd.AddCommand(
		NewDiskPayloadCommand(),
		NewDiskFillCommand(),
	)
	return cmd
}

func NewDiskPayloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-payload <subcommand>",
		Short: "add disk payload",
	}

	cmd.AddCommand(
		NewDiskWritePayloadCommand(),
		NewDiskReadPayloadCommand(),
	)

	return cmd
}

func NewDiskWritePayloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write",
		Short: "write payload",
		Run:   WriteDiskPayloadCommandFunc,
	}

	cmd.Flags().Uint64VarP(&dFlag.Size, "size", "s", 0,
		"'size' specifies how many data will fill in the file path with unit MB.")
	cmd.Flags().StringVarP(&dFlag.Path, "path", "p", "/dev/null",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, payload will write into /dev/null")
	return cmd
}

func WriteDiskPayloadCommandFunc(cmd *cobra.Command, args []string) {
	dFlag.Action = core.DiskWritePayloadAction
	if err := dFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}
	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.DiskPayload(&dFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Fill file %s successfully, uid: %s", dFlag.Path, uid))
}

func NewDiskReadPayloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "read payload",
		Run:   ReadDiskPayloadCommandFunc,
	}

	cmd.Flags().Uint64VarP(&dFlag.Size, "size", "s", 0,
		"'size' specifies how many data will read from the file path with unit MB.")
	cmd.Flags().StringVarP(&dFlag.Path, "path", "p", "/dev/sda",
		"'path' specifies the location to read data.\n"+
			"If path not provided, payload will read from /dev/sda (a disk device)")
	return cmd
}

func ReadDiskPayloadCommandFunc(cmd *cobra.Command, args []string) {
	dFlag.Action = core.DiskReadPayloadAction
	if err := dFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}
	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.DiskPayload(&dFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Fill file %s successfully, uid: %s", dFlag.Path, uid))
}

func NewDiskFillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "fill disk",
		Run:   DiskFillCommandFunc,
	}

	cmd.Flags().Uint64VarP(&dFlag.Size, "size", "s", 0,
		"'size' specifies how many data will fill in the file path with unit MB.")
	cmd.Flags().StringVarP(&dFlag.Path, "path", "p", "",
		"'path' specifies the location to fill data in.\n"+
			"If path not provided, a temp file will be generated and delete after data filled in")

	return cmd
}

func DiskFillCommandFunc(cmd *cobra.Command, args []string) {
	dFlag.Action = core.DiskFillAction
	if err := dFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}
	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.DiskFill(&dFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Fill file %s successfully, uid: %s", dFlag.Path, uid))
}
