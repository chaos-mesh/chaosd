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
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/pingcap/errors"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"

	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type processAttack struct{}

var ProcessAttack AttackType = processAttack{}

func (processAttack) Attack(options core.AttackConfig, _ Environment) error {
	attack := options.(*core.ProcessCommand)

	processes, err := process.Processes()
	if err != nil {
		return errors.WithStack(err)
	}

	notFound := true
	for _, p := range processes {
		pid := int(p.Pid)
		name, err := p.Name()
		fmt.Println(pid, name)
		if attack.Process == strconv.Itoa(pid) || attack.Process == name {
			notFound = false

			err = syscall.Kill(pid, syscall.Signal(attack.Signal))
			if err != nil {
				err = errors.Annotate(err, fmt.Sprintf("kill process with signal %d", attack.Signal))
				return errors.WithStack(err)
			}
			attack.PIDs = append(attack.PIDs, pid)
		}
	}

	if notFound {
		err = errors.Errorf("process %s not found", attack.Process)
		return errors.WithStack(err)
	}

	return nil
}

func (processAttack) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	pcmd := config.(*core.ProcessCommand)
	if pcmd.Signal != int(syscall.SIGSTOP) {
		if pcmd.RecoverCmd == "" {
			return core.ErrNonRecoverableAttack.New("only SIGSTOP process attack and process attack with the recover-cmd are supported to recover")
		}

		rcmd := exec.Command("bash", "-c", pcmd.RecoverCmd)
		if err := rcmd.Start(); err != nil {
			return errors.WithStack(err)
		}

		log.Info("Execute recover-cmd successfully", zap.String("recover-cmd", pcmd.RecoverCmd))

	} else {
		for _, pid := range pcmd.PIDs {
			if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
