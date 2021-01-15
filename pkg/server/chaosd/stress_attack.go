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
	"strings"
	"syscall"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/google/uuid"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func (s *Server) StressAttack(attack *core.StressCommand) (string, error) {
	var err error
	uid := uuid.New().String()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.StressAttack,
		Action:         attack.Action,
		RecoverCommand: attack.String(),
	}); err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), attack.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}

		// use the stressngPid as recover command, and will kill the pid when recover
		if err := s.exp.Update(context.Background(), uid, core.Success, "", attack.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

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
			Options: attack.Options,
		}
	}

	stressorsStr, err := stressors.Normalize()
	if err != nil {
		return "", err
	}
	log.Info("stressors normalize", zap.String("arguments", stressorsStr))

	cmd := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(stressorsStr)...).
		Build()

	// Build will set SysProcAttr.Pdeathsig = syscall.SIGTERM, and so stress-ng will exit while chaosd exit
	// so reset it here
	cmd.Cmd.SysProcAttr = &syscall.SysProcAttr{}

	backgroundProcessManager := bpm.NewBackgroundProcessManager()
	err = backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return "", err
	}
	log.Info("Start stress-ng process successfully", zap.String("command", cmd.String()))

	attack.StressngPid = int32(cmd.Process.Pid)

	return uid, nil
}

func (s *Server) RecoverStressAttack(uid string, attack *core.StressCommand) error {
	proc, err := process.NewProcess(attack.StressngPid)
	if err != nil {
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

	if err := s.exp.Update(context.Background(), uid, core.Destroyed, "", attack.String()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
