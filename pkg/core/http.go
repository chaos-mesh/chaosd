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
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
	"github.com/pingcap/errors"
)

const (
	TargetRequest  tproxyconfig.PodHttpChaosTarget = "Request"
	TargetResponse tproxyconfig.PodHttpChaosTarget = "Response"
)

const (
	HTTPAbortAction = "abort"
	HTTPDelayAction = "delay"
	HTTPFileAction  = "file"
)

var _ AttackConfig = &HTTPAttackConfig{}

type HTTPAttackConfig struct {
	CommonAttackConfig
	Config  tproxyconfig.Config
	process *exec.Cmd
}

func (c HTTPAttackConfig) RecoverData() string {
	data, _ := json.Marshal(c)
	return string(data)
}

type HTTPAttackOption struct {
	CommonAttackConfig
	ProxyPorts []uint                            `json:"proxy_ports,omitempty"`
	Rule       tproxyconfig.PodHttpChaosBaseRule `json:"rules"`
	Path       string                            `json:"path"`
}

func NewHTTPAttackOption() *HTTPAttackOption {
	port := int32(0)
	path := ""
	method := ""
	code := int32(0)
	return &HTTPAttackOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: HTTPAttack,
		},
		ProxyPorts: nil,
		Rule: tproxyconfig.PodHttpChaosBaseRule{
			Target: "",
			Selector: tproxyconfig.PodHttpChaosSelector{
				Port:   &port,
				Path:   &path,
				Method: &method,
				Code:   &code,
			},
		},
	}
}

func (o *HTTPAttackOption) PreProcess() (*HTTPAttackConfig, error) {
	var c tproxyconfig.Config
	switch o.CommonAttackConfig.Action {
	case HTTPAbortAction, HTTPDelayAction:
		if *o.Rule.Selector.Path == "" {
			o.Rule.Selector.Path = nil
		}
		if *o.Rule.Selector.Method == "" {
			o.Rule.Selector.Method = nil
		}
		if *o.Rule.Selector.Code == 0 {
			o.Rule.Selector.Code = nil
		}
		if *o.Rule.Selector.Port == 0 {
			o.Rule.Selector.Port = nil
		}
		switch o.Rule.Target {
		case TargetRequest, TargetResponse:
		default:
			return nil, errors.New("HTTP Attack Target must be Request or Response")
		}
		if o.CommonAttackConfig.Action == HTTPDelayAction {
			o.Rule.Actions.Abort = nil
			_, err := time.ParseDuration(*o.Rule.Actions.Delay)
			if err != nil {
				return nil, errors.Wrapf(err, "HTTP Delay")
			}
		} else {
			o.Rule.Actions.Delay = nil
		}
		ports := make([]uint32, len(o.ProxyPorts))
		for i, port := range o.ProxyPorts {
			ports[i] = uint32(port)
		}
		c.ProxyPorts = ports
		c.Rules = []tproxyconfig.PodHttpChaosBaseRule{o.Rule}
	case HTTPFileAction:
		b, err := os.ReadFile(o.Path)
		if err != nil {
			return nil, errors.Wrap(err, "read HTTP attack config file")
		}

		ext := filepath.Ext(o.Path)
		switch ext {
		case "json":
			err := json.Unmarshal(b, &c)
			if err != nil {
				return nil, errors.Wrap(err, "json unmarshal HTTP attack config file")
			}
		default:
			return nil, errors.Errorf("ext: %s, is not support", ext)
		}
	default:
		return nil, errors.Errorf("unsupported action: %s", o.CommonAttackConfig.Action)
	}
	if len(c.ProxyPorts) == 0 {
		return nil, errors.New("proxy_ports is not an option, you must offer it.")
	}

	return &HTTPAttackConfig{
		CommonAttackConfig: o.CommonAttackConfig,
		Config:             c,
	}, nil
}
