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
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

func (s *Server) ProcessAttack(attack *core.ProcessCommand) (string, error) {
	processes, err := ps.Processes()
	if err != nil {
		return "", errors.WithStack(err)
	}

	uid := uuid.New()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:    uid.String(),
		Status: core.Created,
		Kind:   "process attack",
	}); err != nil {
		return "", errors.WithStack(err)
	}

	for _, p := range processes {
		if attack.Process == strconv.Itoa(p.Pid()) || attack.Process == p.Executable() {
			switch attack.Signal {
			case "KILL":
				syscall.Kill(p.Pid(), syscall.SIGKILL)
			case "TERM":
				syscall.Kill(p.Pid(), syscall.SIGTERM)
			default:
				return "", errors.Errorf("signal %s is not supported", attack.Signal)
			}
		}
	}

	if err := s.exp.Update(context.Background(), uid.String(), core.Success, ""); err != nil {
		return "", errors.WithStack(err)
	}

	return uid.String(), nil
}
