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

	"github.com/mitchellh/go-ps"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type processAttack struct{}

var ProcessAttack AttackType = processAttack{}

func (processAttack) Attack(options core.AttackConfig, _ Environment) error {
	attack := options.(*core.ProcessCommand)

	processes, err := ps.Processes()
	if err != nil {
		return errors.WithStack(err)
	}

	notFound := true
	for _, p := range processes {
		if attack.Process == strconv.Itoa(p.Pid()) || attack.Process == p.Executable() {
			notFound = false

			err = syscall.Kill(p.Pid(), syscall.Signal(attack.Signal))
			if err != nil {
				err = errors.Annotate(err, fmt.Sprintf("kill process with signal %d", attack.Signal))
				return errors.WithStack(err)
			}
			attack.PIDs = append(attack.PIDs, p.Pid())
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
		} else {
			fmt.Printf("Execute recover-cmd %s successfully\n", pcmd.RecoverCmd)
		}

	} else {
		for _, pid := range pcmd.PIDs {
			if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
