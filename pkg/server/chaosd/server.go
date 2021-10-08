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
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"

	"github.com/chaos-mesh/chaosd/pkg/config"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/scheduler"
)

type Server struct {
	expStore     core.ExperimentStore
	ExpRun       core.ExperimentRunStore
	Cron         scheduler.Scheduler
	ipsetRule    core.IPSetRuleStore
	iptablesRule core.IptablesRuleStore
	tcRule       core.TCRuleStore
	conf         *config.Config
	svr          *chaosdaemon.DaemonServer
}

func NewServer(
	conf *config.Config,
	exp core.ExperimentStore,
	expRun core.ExperimentRunStore,
	ipset core.IPSetRuleStore,
	iptables core.IptablesRuleStore,
	tc core.TCRuleStore,
	svr *chaosdaemon.DaemonServer,
	cron scheduler.Scheduler,
) *Server {
	return &Server{
		conf:         conf,
		expStore:     exp,
		Cron:         cron,
		ExpRun:       expRun,
		ipsetRule:    ipset,
		iptablesRule: iptables,
		tcRule:       tc,
		svr:          svr,
	}
}
