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
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/google/uuid"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

const (
	StressAttack = "stress attack"
)

func (s *Server) StressAttack(attack *core.StressCommand) (string, error) {
	var err error
	uid := uuid.New().String()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           StressAttack,
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

	backgroundProcessManager := bpm.NewBackgroundProcessManager()
	err = backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return "", err
	}
	log.Info("Start stress-ng process successfully", zap.String("command", cmd.String()))

	// stress attack will stop while exit
	time.Sleep(attack.Duration)
	kerr := cmd.Process.Kill()
	if kerr != nil {
		log.Error("stress-ng exit failed, please kill it by manual", zap.Error(kerr), zap.Int("process id", cmd.Process.Pid))
	}

	return uid, nil
}

func (s *Server) RecoverStressAttack(uid string, attack *core.ProcessCommand) error {
	// stress attack don't need to recover
	return nil
}
