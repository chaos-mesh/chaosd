// Copyright 2021 Chaos Mesh Authors.
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

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package chaosd

import (
	"os/exec"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

type UnixHost struct{}

var Host HostManager = UnixHost{}

const CmdShutdown = "shutdown"

const CmdReboot = "reboot"

func (h UnixHost) Name() string {
	return "unix"
}

func (h UnixHost) Shutdown() error {
	cmd := exec.Command(CmdShutdown)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
	}
	return err
}

func (h UnixHost) Reboot() error {
	cmd := exec.Command(CmdReboot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
	}
	return err
}
