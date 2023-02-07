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

func NewUserDefinedCommand(uid *string) *cobra.Command {
	options := core.NewUserDefinedOption()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.UserDefinedOption {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "user-defined <subcommand>",
		Short: "user defined attack related commands",
		Run: func(*cobra.Command, []string) {
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(userDefinedAttackFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.AttackCmd, "attack-cmd", "a", "", "the command to be executed when attack")
	cmd.Flags().StringVarP(&options.RecoverCmd, "recover-cmd", "r", "", "the command to be executed when recover")

	return cmd
}

func userDefinedAttackFunc(options *core.UserDefinedOption, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.UserDefinedAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack user defined comamnd successfully, uid: %s", uid))
}
