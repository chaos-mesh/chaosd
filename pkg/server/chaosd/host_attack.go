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
	perr "github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

type HostManager interface {
	Name() string
	Shutdown() error
}

type hostAttack struct{}

var HostAttack AttackType = hostAttack{}

func (hostAttack) Attack(options core.AttackConfig, _ Environment) error {
	if err := Host.Shutdown(); err != nil {
		return perr.WithStack(err)
	}
	return nil
}

func (hostAttack) Recover(exp core.Experiment, _ Environment) error {
	return utils.ErrNonRecoverable(exp.Uid)
}
