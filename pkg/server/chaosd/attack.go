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

// Package chaosd executes fault injection and recovery and manages experiment state.
package chaosd

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type Environment struct {
	//revive:disable-next-line:var-naming Preserve the public attack environment field.
	AttackUid string
	Chaos     *Server
}

type AttackType interface {
	// Attack may add PIDs or original state to options for recovery.
	// For immediate attacks, ExecuteAttack persists the updated recovery data even if Attack returns an error.
	// For scheduled attacks, recovery uses the data from the latest completed Attack call.
	Attack(options core.AttackConfig, env Environment) error
	// Recover can get marshaled options data from experiment and recover it.
	Recover(experiment core.Experiment, env Environment) error
}

func (s *Server) newEnvironment(uid string) Environment {
	return Environment{
		AttackUid: uid,
		Chaos:     s,
	}
}

// ExecuteAttack creates an experiment and runs the attack immediately if options.Cron() is empty.
// Otherwise, it schedules the attack using options.Cron().
func (s *Server) ExecuteAttack(attackType AttackType, options core.AttackConfig, launchMode string) (uid string, err error) {
	uid = options.GetUID()
	if len(uid) == 0 {
		uid = uuid.New().String()
	}

	exp := &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           options.AttackKind(),
		Action:         options.String(),
		RecoverCommand: options.RecoverData(),
		LaunchMode:     launchMode,
	}
	if err = s.expStore.Set(context.Background(), exp); err != nil {
		err = perr.WithStack(err)
		return
	}

	defer func() {
		if err != nil {
			if err := s.expStore.Update(context.Background(), uid, core.Error, err.Error(), options.RecoverData()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}

		var newStatus string
		if len(options.Cron()) > 0 {
			newStatus = core.Scheduled
		} else {
			newStatus = core.Success
		}
		if err := s.expStore.Update(context.Background(), uid, newStatus, "", options.RecoverData()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	env := s.newEnvironment(uid)
	if len(options.Cron()) > 0 {
		attackFunc, recoverFunc := scheduledRunFuncs(attackType, options, exp, env)
		if err = s.Cron.Schedule(exp, options.Cron(), attackFunc, recoverFunc); err != nil {
			err = perr.WithStack(err)
			return
		}
	} else {
		if err = attackType.Attack(options, env); err != nil {
			err = perr.WithStack(err)
			return
		}
	}
	return
}

// scheduledRunFuncs returns the attack and recovery functions for each scheduled run.
// Attack may add PIDs or original state to options, so the recovery function uses
// the data captured after the latest completed attack instead of the data stored
// when the experiment was created.
func scheduledRunFuncs(attackType AttackType, options core.AttackConfig, exp *core.Experiment, env Environment) (func() error, func() error) {
	var (
		lock        sync.Mutex
		recoverData = exp.RecoverCommand
	)

	attackFunc := func() error {
		err := attackType.Attack(options, env)
		lock.Lock()
		recoverData = options.RecoverData()
		lock.Unlock()
		return err
	}

	recoverFunc := func() error {
		lock.Lock()
		data := recoverData
		lock.Unlock()

		// Build a new experiment because exp caches the request command decoded from its creation data.
		return attackType.Recover(core.Experiment{
			ID:             exp.ID,
			Uid:            exp.Uid,
			CreatedAt:      exp.CreatedAt,
			UpdatedAt:      exp.UpdatedAt,
			Status:         exp.Status,
			Message:        exp.Message,
			Kind:           exp.Kind,
			Action:         exp.Action,
			RecoverCommand: data,
			LaunchMode:     exp.LaunchMode,
		}, env)
	}

	return attackFunc, recoverFunc
}
