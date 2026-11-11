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

// RecoverAttack recovers an experiment or cancels its schedule.
// Scheduled experiments must be canceled in the service that owns their cron entry.
// An active scheduled run keeps its existing recovery timer.
func (s *Server) RecoverAttack(uid string) error {
	exp, err := s.expStore.FindByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	if exp == nil {
		return perr.Errorf("experiment %s not found", uid)
	}

	if exp.Status == core.Destroyed {
		return nil
	}

	if exp.Status != core.Success && exp.Status != core.Scheduled {
		return perr.Errorf("can not recover %s experiment", exp.Status)
	}

	attemptRecovery := true
	if exp.Status == core.Scheduled {
		if err = s.Cron.Remove(exp.ID); err != nil {
			return perr.WithMessage(err, "failed to remove scheduled task")
		}
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
		case core.DiskServerAttack:
			attackType = DiskServerAttack
		case core.JVMAttack:
			attackType = JVMAttack
		case core.ClockAttack:
			attackType = ClockAttack
		case core.KafkaAttack:
			attackType = KafkaAttack
		case core.RedisAttack:
			attackType = RedisAttack
		case core.FileAttack:
			attackType = FileAttack
		case core.HTTPAttack:
			attackType = HTTPAttack
		case core.VMAttack:
			attackType = VMAttack
		case core.UserDefinedAttack:
			attackType = UserDefinedAttack
		default:
			return perr.Errorf("chaos experiment kind %s not found", exp.Kind)
		}

		env := s.newEnvironment(uid)
		if err = attackType.Recover(*exp, env); err != nil {
			return perr.WithMessagef(err, "Recover experiment %s failed", uid)
		}
	}

	if err := s.expStore.Update(context.Background(), uid, core.Destroyed, "", exp.RecoverCommand); err != nil {
		return perr.WithStack(err)
	}
	return nil
}

// RecoveryResult is the outcome of recovering one experiment with RecoverAttacks.
type RecoveryResult struct {
	Experiment *core.Experiment
	// Err is nil when the experiment was recovered.
	// It is a core.ErrNonRecoverableAttack error when the fault cannot be recovered.
	Err error
}

// RecoverAttacks recovers every injected experiment, or only those of kind when it is not empty.
// Experiments are recovered from newest to oldest, so faults stacked on the same resource are undone in reverse order.
// A failed recovery does not stop the remaining ones.
// Scheduled experiments are excluded because their schedules belong to the service that created them.
func (s *Server) RecoverAttacks(kind string) ([]RecoveryResult, error) {
	exps, err := s.expStore.ListByConditions(context.Background(), &core.SearchCommand{
		Kind:   kind,
		Status: core.Success,
	})
	if err != nil {
		return nil, perr.WithStack(err)
	}

	results := make([]RecoveryResult, 0, len(exps))
	for _, exp := range exps {
		results = append(results, RecoveryResult{Experiment: exp, Err: s.RecoverAttack(exp.Uid)})
	}
	return results, nil
}
