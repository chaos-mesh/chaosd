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

package main

import (
	"strings"

	_ "github.com/alecthomas/template"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	_ "github.com/swaggo/swag"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/cmd/attack"
	"github.com/chaos-mesh/chaosd/cmd/recover"
	"github.com/chaos-mesh/chaosd/cmd/search"
	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/cmd/version"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

var logLevel string

// CommandFlags are flags that used in all Commands
var rootCmd = &cobra.Command{
	Use:   "chaosd",
	Short: "A command line client to run chaos experiment",
}

func init() {
	cobra.OnInitialize(setLogLevel)
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", "", "the log level of chaosd, the value can be 'debug', 'info', 'warn' and 'error'")

	rootCmd.AddCommand(
		server.NewServerCommand(),
		attack.NewAttackCommand(),
		recover.NewRecoverCommand(),
		search.NewSearchCommand(),
		version.NewVersionCommand(),
	)

	_ = utils.SetRuntimeEnv()
}

func setLogLevel() {
	conf := &log.Config{Level: logLevel}
	lg, r, err := log.InitLogger(conf)
	if err != nil {
		log.Error("fail to init log", zap.Error(err))
		return
	}
	log.ReplaceGlobals(lg, r)

	// only in debug mode print log of go.uber.org/fx
	if strings.ToLower(logLevel) == "debug" {
		utils.PrintFxLog = true
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}
}
