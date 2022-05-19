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

	"github.com/joomcode/errorx"
	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func (s *Server) RecoverAttack(uid string) error {
	exp, err := s.expStore.FindByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	if exp == nil {
		return perr.Errorf("experiment %s not found", uid)
	}

	if exp.Status != core.Success && exp.Status != core.Scheduled {
		return perr.Errorf("can not recover %s experiment", exp.Status)
	}

	attemptRecovery := true
	if exp.Status == core.Scheduled {
		if err = s.Cron.Remove(exp.ID); err != nil {
			return perr.WithMessage(err, "failed to remove scheduled task")
		}
		// it makes sense to not execute recovery for scheduled attacks
		// by their nature, each run would recover on its own after the given duration
		attemptRecovery = false
	}

	if attemptRecovery {
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
		case core.JVMAttack:
			attackType = JVMAttack
		case core.ClockAttack:
			attackType = ClockAttack
		case core.RedisAttack:
			attackType = RedisAttack
		case core.FileAttack:
			attackType = FileAttack
		case core.VMAttack:
			attackType = VMAttack
		case core.UserDefinedAttack:
			attackType = UserDefinedAttack
		default:
			return perr.Errorf("chaos experiment kind %s not found", exp.Kind)
		}

		env := s.newEnvironment(uid)
		if err = attackType.Recover(*exp, env); err != nil {
			if errorx.IsOfType(err, core.ErrNonRecoverableAttack) {
				log.Warn(err.Error(), zap.String("uid", uid), zap.String("kind", exp.Kind))
				return nil
			}
			return perr.WithMessagef(err, "Recover experiment %s failed", uid)
		}
	}

	if err := s.expStore.Update(context.Background(), uid, core.Destroyed, "", exp.RecoverCommand); err != nil {
		return perr.WithStack(err)
	}
	return nil
}
