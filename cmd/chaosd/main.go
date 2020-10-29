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
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	flag "github.com/spf13/pflag"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/container"
	"github.com/chaos-mesh/chaos-daemon/pkg/server"
	"github.com/chaos-mesh/chaos-daemon/pkg/store"
	"github.com/chaos-mesh/chaos-daemon/pkg/version"
)

func main() {
	cfg := config.NewConfig()
	err := cfg.Parse(os.Args[1:])

	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Fatal("parse cmd flags error", zap.Error(err))
	}

	version.PrintVersionInfo("Chaosd Server")
	if cfg.Version {
		os.Exit(0)
	}

	app := fx.New(
		fx.Provide(
			func() *config.Config {
				return cfg
			},
			container.NewCRIClient,
			bpm.NewBackgroundProcessManager,
		),
		store.Module,
		server.Module,
	)
	app.Run()
}
