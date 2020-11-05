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
	"context"
	"strconv"
	"syscall"

	"github.com/google/uuid"
	"github.com/mitchellh/go-ps"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

const (
	ProcessAttack = "process attack"
)

func (s *Server) ProcessAttack(attack *core.ProcessCommand) (string, error) {
	processes, err := ps.Processes()
	if err != nil {
		return "", errors.WithStack(err)
	}

	uid := uuid.New()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid.String(),
		Status:         core.Created,
		Kind:           ProcessAttack,
		RecoverCommand: attack.String(),
	}); err != nil {
		return "", errors.WithStack(err)
	}

	notFound := true
	for _, p := range processes {
		if attack.Process == strconv.Itoa(p.Pid()) || attack.Process == p.Executable() {
			notFound = false
			var kerr error
			switch attack.Signal {
			case int(syscall.SIGKILL):
				kerr = syscall.Kill(p.Pid(), syscall.SIGKILL)
			case int(syscall.SIGTERM):
				kerr = syscall.Kill(p.Pid(), syscall.SIGTERM)
			case int(syscall.SIGSTOP):
				kerr = syscall.Kill(p.Pid(), syscall.SIGSTOP)
			default:
				return "", errors.Errorf("signal %s is not supported", attack.Signal)
			}

			if kerr != nil {
				return "", errors.WithStack(kerr)
			}
			attack.PIDs = append(attack.PIDs, p.Pid())
		}
	}

	if notFound {
		return "", errors.Errorf("process %s not found", attack.Process)
	}

	if err := s.exp.Update(context.Background(), uid.String(), core.Success, "", attack.PIDs); err != nil {
		return "", errors.WithStack(err)
	}

	return uid.String(), nil
}

func (s *Server) RecoverProcessAttack(uid string, attack *core.ProcessCommand) error {
	if attack.Signal != int(syscall.SIGSTOP) {
		return errors.Errorf("chaos experiment %s not supported to recover", uid)
	}

	for _, pid := range attack.PIDs {
		if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
