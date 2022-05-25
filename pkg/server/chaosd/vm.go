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

package chaosd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type vmAttack struct{}

var VMAttack AttackType = vmAttack{}

func (vm vmAttack) Attack(options core.AttackConfig, env Environment) error {
	vmOption, ok := options.(*core.VMOption)
	if !ok {
		return errors.New("the type is not VMOption")
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("virsh destroy %s", vmOption.VMName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}

	return nil
}

func (vmAttack) Recover(exp core.Experiment, _ Environment) error {
	attackConfig, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}

	vmOption, ok := attackConfig.(*core.VMOption)
	if !ok {
		return errors.New("the type is not VMOption")
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("virsh start %s", vmOption.VMName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}

	return nil
}
