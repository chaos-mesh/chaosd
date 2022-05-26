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

func NewHTTPAttackCommand(uid *string) *cobra.Command {
	option := core.NewHTTPAttackOption()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.HTTPAttackOption {
			option.UID = *uid
			return option
		}),
	)

	cmd := &cobra.Command{
		Use:   "http <subcommand>",
		Short: "HTTP attack related commands",
	}

	cmd.AddCommand(
		NewHTTPAbortCommand(dep, option),
		NewHTTPDelayCommand(dep, option),
		NewHTTPConfigCommand(dep, option),
	)

	return cmd
}

func NewHTTPAbortCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort",
		Short: "abort selected HTTP Package",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPAbortAction
			abort := true
			o.Rule.Actions.Abort = &abort
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	setTarget(cmd, o)
	setSelector(cmd, o)
	return cmd
}

func NewHTTPDelayCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delay",
		Short: "delay selected HTTP Package",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPDelayAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	setTarget(cmd, o)
	setSelector(cmd, o)

	delay := ""
	o.Rule.Actions.Delay = &delay
	cmd.Flags().StringVarP(o.Rule.Actions.Delay, "delay time", "d", "", "Delay represents the delay of the target request/response.")
	return cmd
}

func setTarget(cmd *cobra.Command, o *core.HTTPAttackOption) {
	cmd.Flags().UintSliceVarP(&o.ProxyPorts, "proxy-ports", "p", nil,
		"composed with one of the port of HTTP connection, "+
			"we will only attack HTTP connection with port inside proxy_ports")
	cmd.Flags().StringVarP((*string)(&o.Rule.Target), "target", "t", "",
		"HTTP target: Request or Response")
}

func setSelector(cmd *cobra.Command, c *core.HTTPAttackOption) {
	cmd.Flags().Int32Var(c.Rule.Selector.Port, "port", 0, "port is a rule to select server listening on specific port.")
	cmd.Flags().StringVar(c.Rule.Selector.Path, "path", "",
		"Match path of Uri with wildcard matches.")
	cmd.Flags().StringVarP(c.Rule.Selector.Method, "method", "m", "", "HTTP method")
	cmd.Flags().Int32VarP(c.Rule.Selector.Code, "code", "c", 0, "Code is a rule to select target by http status code in response.")
}

func NewHTTPConfigCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "attack with config file",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPConfigAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}
	cmd.Flags().StringVarP(&o.Path, "file path", "p", "", "Config file path.")
	return cmd
}

func processHTTPAttack(o *core.HTTPAttackOption, chaos *chaosd.Server) {
	attackConfig, err := o.PreProcess()
	if err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.HTTPAttack, attackConfig, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("HTTP attack successfully, uid: %s", uid))
}
