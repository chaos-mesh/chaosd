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
	"flag"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pingcap/log"
	"github.com/pkg/errors"

	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/container"
	"github.com/chaos-mesh/chaos-daemon/pkg/server"
	"github.com/chaos-mesh/chaos-daemon/pkg/store"
	"github.com/chaos-mesh/chaos-daemon/pkg/version"
)

func main() {
	// Flushing any buffered log entries
	defer log.Sync() //nolint:errcheck

	version.PrintVersionInfo("Chaosd Server")
	cfg := config.NewConfig()
	err := cfg.Parse(os.Args[1:])

	if cfg.Version {
		os.Exit(0)
	}

	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Fatal("parse cmd flags error", zap.Error(err))
	}

	app := fx.New(
		fx.Provide(
			func() *config.Config {
				return cfg
			},
			container.NewCRIClient,
		),
		store.Module,
		server.Module,
	)
	app.Run()
}
