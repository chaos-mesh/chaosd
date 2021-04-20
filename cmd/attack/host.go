// Copyright 2021 Chaos Mesh Authors.
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

func NewHostAttackCommand() *cobra.Command {
	options := core.NewHostCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.HostCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "host <subcommand>",
		Short: "Host attack related commands",
	}

	cmd.AddCommand(NewHostShutdownCommand(dep, options))

	return cmd
}

func NewHostShutdownCommand(dep fx.Option, options *core.HostCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shutdown",
		Short: "shutdowns system, this action will trigger shutdown of the host machine",

		Run: func(*cobra.Command, []string) {
			fx.New(dep, fx.Invoke(hostAttackF)).Run()
		},
	}

	return cmd
}

func hostAttackF(chaos *chaosd.Server, options *core.HostCommand) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.HostAttack, options)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack host successfully, uid: %s", uid))
}
