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

	"github.com/google/uuid"
	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type Environment struct {
	AttackUid string
	Chaos     *Server
}

type AttackType interface {
	// Attack execute attack with options and env.
	// ExecuteAttack will store the options ahead of Attack be executed
	// and will store options again after Attack be executed.
	// We can also use env.Chaos.expStore to touch the storage of chaosd.
	// But do not update it with your own uid ,
	// because it will be covered after Attack executed with options.
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

// ExecuteAttack creates a new Experiment record and may schedule/execute
// an attack for the given attackType.
// If options.Schedule isn't provided, then the attack is executed immediately.
// Otherwise the attack is scheduled based on the provided schedule spec and duration.
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
		if err = s.Cron.Schedule(
			exp,
			options.Cron(),
			func() error { return attackType.Attack(options, env) },
			func() error { return attackType.Recover(*exp, env) },
		); err != nil {
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
