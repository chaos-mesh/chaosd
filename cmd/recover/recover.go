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

package recover

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type recoverCommand struct {
	uid string
}

func NewRecoverCommand() *cobra.Command {
	options := &recoverCommand{}
	completionCtx := &completionContext{}
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *recoverCommand {
			return options
		}),
		fx.Provide(func() *completionContext {
			return completionCtx
		}),
	)

	cmd := &cobra.Command{
		Use:   "recover UID",
		Short: "Recover a chaos experiment",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			utils.FxNewAppWithoutLog(dep, fx.Invoke(listUid)).Run()
			if completionCtx.err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completionCtx.uids, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				utils.ExitWithMsg(utils.ExitBadArgs, "UID is required")
			}
			options.uid = args[0]
			utils.FxNewAppWithoutLog(dep, fx.Invoke(recoverCommandF)).Run()
		},
	}
	return cmd
}

func recoverCommandF(chaos *chaosd.Server, options *recoverCommand) {
	err := chaos.RecoverAttack(options.uid)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Recover %s successfully", options.uid))
}
