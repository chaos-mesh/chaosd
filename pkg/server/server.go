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
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"

	"github.com/chaos-mesh/chaosd/pkg/crclient"
	"github.com/chaos-mesh/chaosd/pkg/scheduler"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/server/httpserver"
)

var Module = fx.Options(
	fx.Provide(
		provideNIl,
		chaosd.NewServer,
		httpserver.NewServer,
		crclient.NewNodeCRClient,
		os.Getpid,
		chaosdaemon.NewDaemonServerWithCRClient,
		scheduler.NewScheduler,
	),
)

func provideNIl() (prometheus.Registerer, logr.Logger) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger := zapr.NewLogger(zapLogger)
	return nil, logger
}
