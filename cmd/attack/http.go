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
		Short: "Inject HTTP faults or send HTTP requests",
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
		Short: "Abort matching HTTP requests or responses",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPAbortAction
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
		Short: "Delay matching HTTP requests or responses",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPDelayAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	setTarget(cmd, o)
	setSelector(cmd, o)

	cmd.Flags().StringVarP(&o.Delay, "delay-time", "d", "", "Delay duration, e.g. 500ms or 2s")
	return cmd
}

func setTarget(cmd *cobra.Command, o *core.HTTPAttackOption) {
	cmd.Flags().UintSliceVarP(&o.ProxyPorts, "proxy-ports", "p", nil,
		"Required comma-separated list of TCP ports whose HTTP traffic is intercepted")
	cmd.Flags().StringVarP(&o.Target, "target", "t", "",
		"Required HTTP target: Request or Response")
}

func setSelector(cmd *cobra.Command, c *core.HTTPAttackOption) {
	cmd.Flags().Int32Var(&c.Port, "port", 0, "Match the destination TCP port")
	cmd.Flags().StringVar(&c.Path, "path", "",
		"Match the request URI path; supports wildcards")
	cmd.Flags().StringVarP(&c.Method, "method", "m", "", "Match the HTTP method, e.g. GET or POST")
	cmd.Flags().StringVarP(&c.Code, "code", "c", "", "Match the HTTP response status code, e.g. 200 or 503")
}

func NewHTTPConfigCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Apply HTTP fault rules from a JSON configuration file",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPConfigAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}
	cmd.Flags().StringVarP(&o.FilePath, "file-path", "p", "", "Path to the JSON HTTP fault configuration")
	return cmd
}

func NewHTTPRequestCommand(dep fx.Option, o *core.HTTPAttackOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Send HTTP GET requests to a URL",
		Run: func(*cobra.Command, []string) {
			o.Action = core.HTTPRequestAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processHTTPAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&o.URL, "url", "", "", "URL to request")
	cmd.Flags().IntVarP(&o.Count, "count", "c", 1, "Number of requests to send")
	cmd.Flags().BoolVarP(&o.EnableConnPool, "enable-conn-pool", "p", false, "Reuse HTTP connections between requests")
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
