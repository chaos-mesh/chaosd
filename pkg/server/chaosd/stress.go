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
	"errors"
	"io/fs"
	"strings"
	"syscall"

	"github.com/go-logr/zapr"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/pingcap/log"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type stressAttack struct{}

var StressAttack AttackType = stressAttack{}

func (stressAttack) Attack(options core.AttackConfig, _ Environment) (err error) {
	attack := options.(*core.StressCommand)
	stressors := &v1alpha1.Stressors{}
	if attack.Action == core.StressCPUAction {
		stressors.CPUStressor = &v1alpha1.CPUStressor{
			Stressor: v1alpha1.Stressor{
				Workers: attack.Workers,
			},
			Load:    &attack.Load,
			Options: attack.Options,
		}
	} else if attack.Action == core.StressMemAction {
		stressors.MemoryStressor = &v1alpha1.MemoryStressor{
			Stressor: v1alpha1.Stressor{
				Workers: attack.Workers,
			},
			Size:    attack.Size,
			Options: attack.Options,
		}
	}

	errs := stressors.Validate(nil, field.NewPath("stressors"))
	if len(errs) > 0 {
		return errors.New(errs.ToAggregate().Error())
	}

	stressorsStr, _, err := stressors.Normalize()
	if err != nil {
		return
	}
	log.Info("stressors normalize", zap.String("arguments", stressorsStr))

	cmd := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(stressorsStr)...).
		Build(context.Background())

	// Build will set SysProcAttr.Pdeathsig = syscall.SIGTERM, and so stress-ng will exit while chaosd exit
	// so reset it here
	cmd.Cmd.SysProcAttr = &syscall.SysProcAttr{}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger := zapr.NewLogger(zapLogger)
	backgroundProcessManager := bpm.StartBackgroundProcessManager(nil, logger)
	_, err = backgroundProcessManager.StartProcess(context.Background(), cmd)
	if err != nil {
		return
	}

	attack.StressngPid = int32(cmd.Process.Pid)
	log.Info("Start stress-ng process successfully", zap.String("command", cmd.String()), zap.Int32("Pid", attack.StressngPid))

	return nil
}

func (stressAttack) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	attack := config.(*core.StressCommand)
	proc, err := process.NewProcess(attack.StressngPid)
	if err != nil {
		log.Warn("Failed to get stress-ng process", zap.Error(err))
		if errors.Is(err, process.ErrorProcessNotRunning) || errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return err
	}

	procName, err := proc.Name()
	if err != nil {
		return err
	}

	if !strings.Contains(procName, "stress-ng") {
		log.Warn("the process is not stress-ng, maybe it is killed by manual")
		return nil
	}

	if err := proc.Kill(); err != nil {
		log.Error("the stress-ng process kill failed", zap.Error(err))
		return err
	}

	return nil
}
