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

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var pFlag core.ProcessCommand

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
		Use:   "kill",
		Short: "kill process, default signal 9",
		Run:   processKillCommandFunc,
	}

	cmd.Flags().StringVarP(&pFlag.Process, "process", "p", "", "The process name or the process ID")
	cmd.Flags().IntVarP(&pFlag.Signal, "single", "s", 9, "The signal number to send")

	return cmd
}

func NewProcessStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "stop process, this action will stop the process with SIGSTOP",

		Run: processStopCommandFunc,
	}

	cmd.Flags().StringVarP(&pFlag.Process, "process", "p", "", "The process name or the process ID")
	pFlag.Signal = int(syscall.SIGSTOP)

	return cmd
}

func processKillCommandFunc(cmd *cobra.Command, args []string) {
	processAttackF(cmd, &pFlag)
}

func processStopCommandFunc(cmd *cobra.Command, args []string) {
	processAttackF(cmd, &pFlag)
}

func processAttackF(cmd *cobra.Command, f *core.ProcessCommand) {
	if err := pFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.ProcessAttack(f)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack process %s successfully, uid: %s", f.Process, uid))
}
