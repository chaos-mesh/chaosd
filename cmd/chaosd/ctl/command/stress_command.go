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
	//"syscall"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var sFlag core.StressCommand

func NewStressAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress <subcommand>",
		Short: "stress attack related commands",
	}

	cmd.AddCommand(
		NewStressCPUCommand(),
		NewStressMemCommand(),
	)

	return cmd
}

func NewStressCPUCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cpu [options]",
		Short: "kill process, default signal 9",
		Run:   stressCPUCommandFunc,
	}

	sFlag.Action = core.StressCPUAction
	cmd.Flags().IntVarP(&sFlag.Load, "load", "l", 0, "Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100 is full loading.")
	cmd.Flags().IntVarP(&sFlag.Workers, "workers", "w", 1, "Workers specifies N workers to apply the stressor.")
	cmd.Flags().StringSliceVarP(&sFlag.Options, "options", "o", []string{}, "extend stress-ng options.")
	//cmd.Flags().StringVarP(&pFlag.Process, "process", "p", "", "The process name or the process ID")
	//cmd.Flags().IntVarP(&pFlag.Signal, "single", "s", 9, "The signal number to send")

	return cmd
}

func NewStressMemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mem [options]",
		Short: "stop process, this action will stop the process with SIGSTOP",

		Run: stressMemCommandFunc,
	}

	sFlag.Action = core.StressMemAction
	cmd.Flags().IntVarP(&sFlag.Workers, "workers", "w", 1, "Workers specifies N workers to apply the stressor.")
	cmd.Flags().StringSliceVarP(&sFlag.Options, "options", "o", []string{}, "extend stress-ng options.")
	//cmd.Flags().StringVarP(&pFlag.Process, "process", "p", "", "The process name or the process ID")
	//pFlag.Signal = int(syscall.SIGSTOP)

	return cmd
}

func stressCPUCommandFunc(cmd *cobra.Command, args []string) {
	stressAttackF(cmd, &sFlag)
}

func stressMemCommandFunc(cmd *cobra.Command, args []string) {
	stressAttackF(cmd, &sFlag)
}

func stressAttackF(cmd *cobra.Command, s *core.StressCommand) {
	if err := sFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.StressAttack(s)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack stress %s successfully, uid: %s", s.Action, uid))
}
