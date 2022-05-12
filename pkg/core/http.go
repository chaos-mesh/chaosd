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

package core

import (
	"encoding/json"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

var _ AttackConfig = &HTTPAttackConfig{}

const (
	TargetRequest  tproxyconfig.PodHttpChaosTarget = "Request"
	TargetResponse tproxyconfig.PodHttpChaosTarget = "Response"
)

const (
	HTTPAbortAction = "abort"
	HTTPDelayAction = "delay"
)

type HTTPAttackConfig struct {
	CommonAttackConfig
	ProxyPorts []uint                            `json:"proxy_ports,omitempty"`
	Rule       tproxyconfig.PodHttpChaosBaseRule `json:"rules"`
}

func (c HTTPAttackConfig) RecoverData() string {
	data, _ := json.Marshal(c)
	return string(data)
}
