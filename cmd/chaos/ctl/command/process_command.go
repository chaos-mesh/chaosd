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
	"syscall"

	"github.com/pingcap/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

var (
	process string
)

func NewProcessAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process <subcommand>",
		Short: "Process attack related commands",
	}

	cmd.AddCommand(
		NewProcessKillCommand(),
		NewProcessStopCommand(),
	)

	return cmd
}

func NewProcessKillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill [options]",
		Short: "kill process, default signal SIGKILL",
		Run:   processKillCommandFunc,
	}

	cmd.Flags().StringVarP(&process, "process", "p", "", "The process name or the process ID")

	return cmd
}

func NewProcessStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [options]",
		Short: "stop process, this action will stop the process with SIGSTOP",

		Run: processStopCommandFunc,
	}

	cmd.Flags().StringVarP(&process, "process", "p", "", "The process name or the process ID")

	return cmd
}

func processKillCommandFunc(cmd *cobra.Command, args []string) {
	processAttackF(cmd, syscall.SIGKILL)
}

func processStopCommandFunc(cmd *cobra.Command, args []string) {
	processAttackF(cmd, syscall.SIGSTOP)
}

func processAttackF(cmd *cobra.Command, sig syscall.Signal) {
	if len(process) == 0 {
		ExitWithError(ExitBadArgs, errors.New("process not provided"))
	}

	cli := mustClientFromCmd(cmd)

	resp, apiErr, err := cli.CreateProcessAttack(&core.ProcessCommand{
		Process: process,
		Signal:  sig,
	})

	if err != nil {
		ExitWithError(ExitError, err)
	}

	if apiErr != nil {
		ExitWithMsg(ExitError, fmt.Sprintf("Failed to attack process %s, %s", process, apiErr.Message))
	}

	if resp.Status != 200 {
		ExitWithMsg(ExitError, fmt.Sprintf("Failed to attack process %s, %s", process, resp.Message))
	}

	NormalExit(fmt.Sprintf("Attack process %s successfully, uid: %s", process, resp.UID))
}
