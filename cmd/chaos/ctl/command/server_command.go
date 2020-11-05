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

package command

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/container"
	"github.com/chaos-mesh/chaos-daemon/pkg/server"
	"github.com/chaos-mesh/chaos-daemon/pkg/store"
	"github.com/chaos-mesh/chaos-daemon/pkg/version"
)

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server <option>",
		Short: "Run Chaosd Server",
		Run:   serverCommandFunc,
	}

	cmd.Flags().IntVarP(&conf.ListenPort, "port", "p", 31767, "listen port of the Chaosd Server")
	cmd.Flags().StringVarP(&conf.ListenHost, "host", "a", "0.0.0.0", "listen host of the Chaosd Server")
	cmd.Flags().StringVarP(&conf.Runtime, "runtime", "r", "docker", "current container runtime")
	cmd.Flags().BoolVar(&conf.EnablePprof, "enable-pprof", true, "enable pprof")
	cmd.Flags().IntVar(&conf.PprofPort, "pprof-port", 31766, "listen port of the pprof server")
	cmd.Flags().StringVarP(&conf.Platform, "platform", "f", "local", "platform to deploy, default: local, supported platform: local, kubernetes")

	return cmd
}

var conf = config.Config{
	Platform: config.LocalPlatform,
	Runtime:  "docker",
}

func serverCommandFunc(cmd *cobra.Command, args []string) {
	if err := conf.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	version.PrintVersionInfo("Chaosd Server")

	app := fx.New(
		fx.Provide(
			func() *config.Config {
				return &conf
			},
			container.NewCRIClient,
			bpm.NewBackgroundProcessManager,
		),
		store.Module,
		server.Module,
	)
	app.Run()
}
