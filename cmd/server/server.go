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

package server

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/pkg/config"
	"github.com/chaos-mesh/chaosd/pkg/server/httpserver"
	"github.com/chaos-mesh/chaosd/pkg/utils"
	"github.com/chaos-mesh/chaosd/pkg/version"
)

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server <option>",
		Short: "Run the chaosd HTTP API server",
		Run:   serverCommandFunc,
	}

	cmd.Flags().IntVarP(&conf.ListenPort, "port", "p", 31767, "HTTP API listening port")
	cmd.Flags().IntVar(&conf.ListenHttpsPort, "https-port", 31768, "HTTPS API listening port; requires --cert and --key")
	cmd.Flags().StringVarP(&conf.ListenHost, "host", "a", "0.0.0.0", "Address to bind the API server to")
	cmd.Flags().StringVar(&conf.SSLCertFile, "cert", "", "Path to the PEM server certificate; requires --key")
	cmd.Flags().StringVar(&conf.SSLKeyFile, "key", "", "Path to the PEM server private key; requires --cert")
	cmd.Flags().StringVar(&conf.SSLClientCAFile, "CA", "", "Path to PEM CA certificates for verifying client certificates; requires HTTPS")
	cmd.Flags().StringVar(&conf.ServerName, "server-name", "chaosd.chaos-mesh.org", "TLS server name setting; not used to verify client certificate hostnames")
	cmd.Flags().StringVarP(&conf.Runtime, "runtime", "r", "docker", "Container runtime setting: docker or containerd")
	cmd.Flags().BoolVar(&conf.EnablePprof, "enable-pprof", true, "Enable the pprof HTTP profiling server")
	cmd.Flags().IntVar(&conf.PprofPort, "pprof-port", 31766, "pprof HTTP listening port")
	cmd.Flags().StringVarP(&conf.Platform, "platform", "f", "local", "Deployment platform: local or kubernetes; the API server starts only in local mode")

	return cmd
}

var conf = config.Config{
	Platform: config.LocalPlatform,
	Runtime:  "docker",
}

func serverCommandFunc(_ *cobra.Command, _ []string) {
	if err := conf.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	version.PrintVersionInfo("Chaosd Server")

	app := utils.FxNewAppWithoutLog(
		Module,
		fx.Invoke(httpserver.Register),
	)
	app.Run()
}
