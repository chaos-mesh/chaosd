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

package chaosd

import (
	"context"

	perr "github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func (s *Server) RecoverAttack(uid string) error {
	exp, err := s.exp.FindByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	if exp == nil {
		return perr.Errorf("experiment %s not found", uid)
	}

	if exp.Status != core.Success || exp.Status != core.Scheduled {
		return perr.Errorf("can not recover %s experiment", exp.Status)
	}

	if len(exp.Cron) > 0 {
		if err = s.Cron.Remove(exp.ID); err != nil {
			return perr.WithMessage(err, "failed to remove scheduled task")
		}
	}

	var attackType AttackType
	switch exp.Kind {
	case core.ProcessAttack:
		attackType = ProcessAttack
	case core.NetworkAttack:
		attackType = NetworkAttack
	case core.HostAttack:
		attackType = HostAttack
	case core.StressAttack:
		attackType = StressAttack
	case core.DiskAttack:
		attackType = DiskAttack
	default:
		return perr.Errorf("chaos experiment kind %s not found", exp.Kind)
	}

	env := s.newEnvironment(uid)
	if err = attackType.Recover(*exp, env); err != nil {
		return perr.WithMessagef(err, "Recover experiment %s failed, %s", uid)
	}

	if err := s.exp.Update(context.Background(), uid, core.Destroyed, "", exp.RecoverCommand); err != nil {
		return perr.WithStack(err)
	}
	return nil
}
