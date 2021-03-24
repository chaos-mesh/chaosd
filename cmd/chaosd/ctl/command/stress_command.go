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
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

var stFlag core.StressCommand

func NewStressAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress <subcommand>",
		Short: "Stress attack related commands",
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
		Short: "continuously stress CPU out",
		Run:   stressCPUCommandFunc,
	}

	cmd.Flags().IntVarP(&stFlag.Load, "load", "l", 10, "Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100 is full loading.")
	cmd.Flags().IntVarP(&stFlag.Workers, "workers", "w", 1, "Workers specifies N workers to apply the stressor.")
	cmd.Flags().StringSliceVarP(&stFlag.Options, "options", "o", []string{}, "extend stress-ng options.")
	commonFlags(cmd, &stFlag.CommonAttackConfig)

	return cmd
}

func NewStressMemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mem [options]",
		Short: "continuously stress virtual memory out",

		Run: stressMemCommandFunc,
	}

	cmd.Flags().IntVarP(&stFlag.Workers, "workers", "w", 1, "Workers specifies N workers to apply the stressor.")
	cmd.Flags().StringVarP(&stFlag.Size, "size", "s", "", "Size specifies N bytes consumed per vm worker, default is the total available memory. One can specify the size as % of total available memory or in units of B, KB/KiB, MB/MiB, GB/GiB, TB/TiB..")
	cmd.Flags().StringSliceVarP(&stFlag.Options, "options", "o", []string{}, "extend stress-ng options.")

	return cmd
}

func stressCPUCommandFunc(cmd *cobra.Command, args []string) {
	stFlag.Action = core.StressCPUAction
	stressAttackF(cmd, &stFlag)
}

func stressMemCommandFunc(cmd *cobra.Command, args []string) {
	stFlag.Action = core.StressMemAction
	stressAttackF(cmd, &stFlag)
}

func stressAttackF(cmd *cobra.Command, s *core.StressCommand) {
	if err := stFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.ProcessAttack(chaosd.StressAttack, s)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack stress %s successfully, uid: %s", s.Action, uid))
}
