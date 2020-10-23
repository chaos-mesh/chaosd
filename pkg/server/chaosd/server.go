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

package chaosd

import (
	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/container"
	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

type Server struct {
	exp                      core.ExperimentStore
	conf                     config.Config
	criCli                   container.CRIClient
	backgroundProcessManager bpm.BackgroundProcessManager
}

func NewServer(exp core.ExperimentStore, cli container.CRIClient, bpm bpm.BackgroundProcessManager) *Server {
	return &Server{
		exp:                      exp,
		criCli:                   cli,
		backgroundProcessManager: bpm,
	}
}
