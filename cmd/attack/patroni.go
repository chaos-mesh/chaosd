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
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func NewPatroniAttackCommand(uid *string) *cobra.Command {
	options := core.NewPatroniCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.PatroniCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "patroni <subcommand>",
		Short: "Patroni attack related commands",
	}

	cmd.AddCommand(
		NewPatroniSwitchoverCommand(dep, options),
		NewPatroniFailoverCommand(dep, options),
	)

	cmd.PersistentFlags().StringVarP(&options.User, "user", "u", "patroni", "patroni cluster user")
	cmd.PersistentFlags().StringVar(&options.Password, "password", "p", "patroni cluster password")

	return cmd
}

func NewPatroniSwitchoverCommand(dep fx.Option, options *core.PatroniCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switchover",
		Short: "exec switchover, default without another attack. Warning! Command is not recover!",
		Run: func(*cobra.Command, []string) {
			options.Action = core.SwitchoverAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(PatroniAttackF)).Run()
		},
	}
	cmd.Flags().StringVarP(&options.Address, "address", "a", "", "patroni cluster address, any of available hosts")
	cmd.Flags().StringVarP(&options.Candidate, "candidate", "c", "", "switchover candidate, default random unit for replicas")
	cmd.Flags().StringVarP(&options.Scheduled_at, "scheduled_at", "d", fmt.Sprintln(time.Now().Add(time.Second*60).Format(time.RFC3339)), "scheduled switchover, default now()+1 minute")

	return cmd
}

func NewPatroniFailoverCommand(dep fx.Option, options *core.PatroniCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "failover",
		Short: "exec failover, default without another attack",
		Run: func(*cobra.Command, []string) {
			options.Action = core.FailoverAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(PatroniAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Address, "address", "a", "", "patroni cluster address, any of available hosts")
	cmd.Flags().StringVarP(&options.Candidate, "leader", "c", "", "failover new leader, default random unit for replicas")
	return cmd
}

func PatroniAttackF(options *core.PatroniCommand, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.PatroniAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack %s successfully to patroni address %s, uid: %s", options.Action, options.Address, uid))
}
