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
		NewHTTPRequestCommand(dep, option),
	)

	return cmd
}

func NewHTTPAbortCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort",
		Short: "abort selected HTTP connection",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPAbortAction
			o.Abort = true
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

	cmd.Flags().StringVarP(&o.Delay, "delay time", "d", "", "Delay represents the delay of the target request/response.")
	return cmd
}

func setTarget(cmd *cobra.Command, o *core.HTTPAttackOption) {
	cmd.Flags().UintSliceVarP(&o.ProxyPorts, "proxy-ports", "p", nil,
		"composed with one of the port of HTTP connection, "+
			"we will only attack HTTP connection with port inside proxy_ports")
	cmd.Flags().StringVarP(&o.Target, "target", "t", "",
		"HTTP target: Request or Response")
}

func setSelector(cmd *cobra.Command, c *core.HTTPAttackOption) {
	cmd.Flags().Int32Var(&c.Port, "port", 0, "The TCP port that the target service listens on.")
	cmd.Flags().StringVar(&c.Path, "path", "",
		"Match path of Uri with wildcard matches.")
	cmd.Flags().StringVarP(&c.Method, "method", "m", "", "HTTP method")
	cmd.Flags().StringVarP(&c.Code, "code", "c", "", "Code is a rule to select target by http status code in response.")
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
	cmd.Flags().StringVarP(&o.FilePath, "file path", "p", "", "Config file path.")
	return cmd
}

func NewHTTPRequestCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "request specific URL",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPRequestAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&o.HTTPRequestConfig.URL, "url", "", "", "Request to send")
	cmd.Flags().IntVarP(&o.HTTPRequestConfig.Count, "count", "c", 1, "Number of requests to send")
	cmd.Flags().BoolVarP(&o.HTTPRequestConfig.EnableConnPool, "enable-conn-pool", "p", false, "Enable connection pool")
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
