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

func NewStressAttackCommand(uid *string) *cobra.Command {
	options := core.NewStressCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.StressCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "stress <subcommand>",
		Short: "Generate CPU or memory stress",
	}

	cmd.AddCommand(
		NewStressCPUCommand(dep, options),
		NewStressMemCommand(dep, options),
	)

	return cmd
}

func NewStressCPUCommand(dep fx.Option, options *core.StressCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cpu [options]",
		Short: "Generate continuous CPU load",
		Run: func(*cobra.Command, []string) {
			options.Action = core.StressCPUAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(stressAttackF)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.Load, "load", "l", 10, "CPU load percentage per worker, from 0 to 100")
	cmd.Flags().IntVarP(&options.Workers, "workers", "w", 1, "Number of stress workers")
	cmd.Flags().StringSliceVarP(&options.Options, "options", "o", []string{}, "Additional stress-ng options, separated by commas")

	return cmd
}

func NewStressMemCommand(dep fx.Option, options *core.StressCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mem [options]",
		Short: "Allocate and hold memory",
		Run: func(*cobra.Command, []string) {
			options.Action = core.StressMemAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(stressAttackF)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.Workers, "workers", "w", 1, "Number of stress workers")
	cmd.Flags().StringVarP(&options.Size, "size", "s", "", "Total memory to allocate across workers, e.g. 100MiB or 25% of total memory")
	cmd.Flags().StringSliceVarP(&options.Options, "options", "o", []string{}, "Additional memStress options, separated by commas")

	return cmd
}

func stressAttackF(chaos *chaosd.Server, options *core.StressCommand) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.StressAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack stress %s successfully, uid: %s", options.Action, uid))
}
