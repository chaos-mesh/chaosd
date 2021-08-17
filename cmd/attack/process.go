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

	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func NewProcessAttackCommand(uid *string) *cobra.Command {
	options := core.NewProcessCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.ProcessCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "process <subcommand>",
		Short: "Process attack related commands",
	}

	cmd.AddCommand(
		NewProcessKillCommand(dep, options),
		NewProcessStopCommand(dep, options),
	)

	return cmd
}

func NewProcessKillCommand(dep fx.Option, options *core.ProcessCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill",
		Short: "kill process, default signal 9",
		Run: func(*cobra.Command, []string) {
			options.Action = core.ProcessKillAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Process, "process", "p", "", "The process name or the process ID")
	cmd.Flags().IntVarP(&options.Signal, "signal", "s", 9, "The signal number to send")

	return cmd
}

func NewProcessStopCommand(dep fx.Option, options *core.ProcessCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "stop process, this action will stop the process with SIGSTOP",
		Run: func(*cobra.Command, []string) {
			options.Signal = int(syscall.SIGSTOP)
			options.Action = core.ProcessStopAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Process, "process", "p", "", "The process name or the process ID")

	return cmd
}

func processAttackF(options *core.ProcessCommand, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.ProcessAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack process %s successfully, uid: %s", options.Process, uid))
}
