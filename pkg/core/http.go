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
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
	"github.com/pingcap/errors"
)

const (
	TargetRequest  tproxyconfig.PodHttpChaosTarget = "Request"
	TargetResponse tproxyconfig.PodHttpChaosTarget = "Response"
)

const (
	HTTPAbortAction   = "abort"
	HTTPDelayAction   = "delay"
	HTTPConfigAction  = "config"
	HTTPRequestAction = "request"
)

var _ AttackConfig = &HTTPAttackConfig{}

type HTTPAttackConfig struct {
	CommonAttackConfig
	Config   tproxyconfig.Config
	ProxyPID int

	Logger logr.Logger

	HTTPRequestConfig
}

func (c HTTPAttackConfig) RecoverData() string {
	data, _ := json.Marshal(c)
	return string(data)
}

type HTTPAttackOption struct {
	CommonAttackConfig

	ProxyPorts []uint `json:"proxy_ports"`
	Target     string `json:"target"`
	Port       int32  `json:"port,omitempty"`
	Path       string `json:"path,omitempty"`
	Method     string `json:"method,omitempty"`
	Code       string `json:"code,omitempty"`
	Abort      bool   `json:"abort"`
	Delay      string `json:"delay"`

	FilePath string `json:"file_path,omitempty"`

	HTTPRequestConfig `json:",inline"`
}

type HTTPRequestConfig struct {
	// used for HTTP request, now only support GET
	URL            string `json:"url,omitempty"`
	EnableConnPool bool   `json:"enable-conn-pool,omitempty"`
	Count          int    `json:"count,omitempty"`
}

func NewHTTPAttackOption() *HTTPAttackOption {
	return &HTTPAttackOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: HTTPAttack,
		},
	}
}

func (o *HTTPAttackOption) PreProcess() (*HTTPAttackConfig, error) {
	var c tproxyconfig.Config
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	logger := zapr.NewLogger(zapLogger).WithName("HTTP Attack")
	switch o.CommonAttackConfig.Action {
	case HTTPAbortAction, HTTPDelayAction:
		switch o.Target {
		case string(TargetRequest), string(TargetResponse):
		default:
			return nil, errors.New("HTTP Attack Target must be Request or Response")
		}
		rule := tproxyconfig.PodHttpChaosBaseRule{
			Target:   tproxyconfig.PodHttpChaosTarget(o.Target),
			Selector: tproxyconfig.PodHttpChaosSelector{},
			Actions:  tproxyconfig.PodHttpChaosActions{},
		}
		if o.Path != "" {
			rule.Selector.Path = &o.Path
		}
		if o.Method != "" {
			rule.Selector.Method = &o.Method
		}
		if o.Code != "" {
			code, err := strconv.Atoi(o.Code)
			if err != nil {
				return nil, errors.Wrapf(err, "parsing %v", o)
			}
			codeI32 := int32(code)
			rule.Selector.Code = &codeI32
		}
		if o.Port != 0 {
			rule.Selector.Port = &o.Port
		}
		rule.Actions.Abort = &o.Abort
		if o.CommonAttackConfig.Action == HTTPDelayAction {
			rule.Actions.Abort = nil
			_, err := time.ParseDuration(o.Delay)
			if err != nil {
				return nil, errors.Wrapf(err, "HTTP Delay")
			}
			rule.Actions.Delay = &o.Delay
		}
		ports := make([]uint32, len(o.ProxyPorts))
		for i, port := range o.ProxyPorts {
			ports[i] = uint32(port)
		}
		c.ProxyPorts = ports
		c.Rules = []tproxyconfig.PodHttpChaosBaseRule{rule}
	case HTTPConfigAction:
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
	case HTTPRequestAction:
		if o.URL == "" {
			return nil, errors.New("URL is required")
		}
	default:
		return nil, errors.Errorf("unsupported action: %s", o.CommonAttackConfig.Action)
	}
	if len(c.ProxyPorts) == 0 && o.CommonAttackConfig.Action != HTTPRequestAction {
		return nil, errors.New("proxy_ports is not an option, you must offer it")
	}

	return &HTTPAttackConfig{
		CommonAttackConfig: o.CommonAttackConfig,
		Config:             c,
		Logger:             logger,
		HTTPRequestConfig:  o.HTTPRequestConfig,
	}, nil
}
