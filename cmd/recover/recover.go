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
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type recoverCommand struct {
	uid string
}

func NewRecoverCommand() *cobra.Command {
	options := &recoverCommand{}
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *recoverCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:               "recover UID",
		Short:             "Recover a chaos experiment",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeUid,
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

func completeUid(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completionCtx := newCompletionCtx()
	completionDep := fx.Options(
		server.Module,
		fx.Provide(func() *completionContext {
			return completionCtx
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := utils.FxNewAppWithoutLog(completionDep, fx.Invoke(listUid)).Start(ctx); err != nil {
			log.Error(errors.Wrap(err, "start application").Error())
		}
	}()
	var uids []string
	for {
		select {
		case uid := <-completionCtx.uids:
			if len(uid) == 0 {
				return uids, cobra.ShellCompDirectiveNoFileComp
			}
			uids = append(uids, uid)
		case err := <-completionCtx.err:
			log.Error(err.Error())
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}
