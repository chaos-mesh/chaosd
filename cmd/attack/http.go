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
	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewHTTPAttackCommand(uid *string) *cobra.Command {
	config := &core.HTTPAttackConfig{}
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.HTTPAttackConfig {
			config.UID = *uid
			return config
		}),
	)

	cmd := &cobra.Command{
		Use:   "http <subcommand>",
		Short: "HTTP attack related commands",
	}

	cmd.AddCommand(
		NewHTTPAbortCommand(dep, config),
		NewHTTPDelayCommand(dep, config),
	)

	return cmd
}

func NewHTTPAbortCommand(dep fx.Option, c *core.HTTPAttackConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort",
		Short: "abort selected HTTP Package",
		Run: func(*cobra.Command, []string) {
			c.Action = core.HTTPAbortAction
			abort := true
			c.Rule.Actions.Abort = &abort
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	cmdTarget(cmd, c)
	cmdSelector(cmd, c)
	return cmd
}

func NewHTTPDelayCommand(dep fx.Option, c *core.HTTPAttackConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delay",
		Short: "delay selected HTTP Package",
		Run: func(*cobra.Command, []string) {
			c.Action = core.HTTPAbortAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	cmdTarget(cmd, c)
	cmdSelector(cmd, c)

	delay := ""
	c.Rule.Actions.Delay = &delay
	cmd.Flags().StringVarP(c.Rule.Actions.Delay, "delay time", "d", "", "Delay represents the delay of the target request/response.")
	return cmd
}

func cmdTarget(cmd *cobra.Command, c *core.HTTPAttackConfig) {
	cmd.Flags().UintSliceVarP(&c.ProxyPorts, "proxy_ports", "p", nil,
		"composed with one of the port of HTTP connection, "+
			"we will only attack HTTP connection with port inside proxy_ports")
	cmd.Flags().StringVarP((*string)(&c.Rule.Target), "target", "t", "Request",
		"HTTP target: Request or Response")
}

func cmdSelector(cmd *cobra.Command, c *core.HTTPAttackConfig) {
	port := int32(0)
	c.Rule.Selector.Port = &port
	cmd.Flags().Int32Var(c.Rule.Selector.Port, "port", 0, "port is a rule to select server listening on specific port.")
	path := ""
	c.Rule.Selector.Path = &path
	cmd.Flags().StringVar(c.Rule.Selector.Path, "path", "",
		"Mathc path of Uri with wildcard matches.")
	meth := ""
	c.Rule.Selector.Method = &meth
	cmd.Flags().StringVarP(c.Rule.Selector.Method, "method", "m", "", "HTTP method")
	code := int32(0)
	c.Rule.Selector.Code = &code
	cmd.Flags().Int32VarP(c.Rule.Selector.Code, "code", "c", 0, "Code is a rule to select target by http status code in response.")
}

func processHTTPAttack(c *core.HTTPAttackConfig, chaos *chaosd.Server) {
	utils.NormalExit(fmt.Sprintf("%v,%v\n", c, chaos))
}
