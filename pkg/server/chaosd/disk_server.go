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
	"fmt"
	"os"
	"time"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	pkgUtils "github.com/chaos-mesh/chaosd/pkg/utils"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type diskServerAttack struct{}

var DiskServerAttack AttackType = diskServerAttack{}

func handleDiskServerOutput(output []byte, err error, _ chan interface{}) {
	if err != nil {
		log.Error(string(output), zap.Error(err))
	}
	log.Info(string(output))
}

func (diskServerAttack) Attack(options core.AttackConfig, env Environment) error {
	err := ApplyDiskServerAttack(options, env)
	if err != nil {
		return err
	}
	return nil
}

type OutputHandler struct {
	StdoutHandler func([]byte, error, chan interface{})
	OutputChan    chan interface{}
}

func NewOutputHandler(
	handler func([]byte, error, chan interface{}),
	outputChan chan interface{}) *OutputHandler {
	return &OutputHandler{
		StdoutHandler: handler,
		OutputChan:    outputChan,
	}
}

func getPoolSize(attackConf *core.DiskAttackConfig) int {
	poolSize := 1
	if attackConf.DdOptions != nil && len(*attackConf.DdOptions) > 0 {
		poolSize = len(*attackConf.DdOptions)
	}
	return poolSize
}

func fillDisk(
	attackConf *core.DiskAttackConfig,
	cmdPool *pkgUtils.CommandPools,
	outputHandler *OutputHandler) {
	if attackConf.FAllocateOption != nil {
		name, args := core.FAllocateCommand.GetCmdArgs(*attackConf.FAllocateOption)
		runner := pkgUtils.NewCommandRunner(name, args).
			WithOutputHandler(outputHandler.StdoutHandler, outputHandler.OutputChan)
		cmdPool.Start(runner)
		return
	}

	for _, DdOption := range *attackConf.DdOptions {
		name, args := core.DdCommand.GetCmdArgs(DdOption)
		runner := pkgUtils.NewCommandRunner(name, args).
			WithOutputHandler(outputHandler.StdoutHandler, outputHandler.OutputChan)
		cmdPool.Start(runner)
	}
	return
}

func getDeadline(options core.AttackConfig) *time.Time {
	duration, _ := options.ScheduleDuration()
	if duration != nil {
		deadline := time.Now().Add(*duration)
		return &deadline
	}
	return nil
}

func applyPayload(
	attackConf *core.DiskAttackConfig,
	cmdPool *pkgUtils.CommandPools,
	outputHandler *OutputHandler) {
	if len(*attackConf.DdOptions) == 0 {
		return
	}
	rest := (*attackConf.DdOptions)[len(*attackConf.DdOptions)-1]
	*attackConf.DdOptions = (*attackConf.DdOptions)[:len(*attackConf.DdOptions)-1]
	name, args := core.DdCommand.GetCmdArgs(rest)
	runner := pkgUtils.NewCommandRunner(name, args).
		WithOutputHandler(outputHandler.StdoutHandler, outputHandler.OutputChan)
	cmdPool.Start(runner)

	for _, ddOpt := range *attackConf.DdOptions {
		name, args := core.DdCommand.GetCmdArgs(ddOpt)
		runner := pkgUtils.NewCommandRunner(name, args).
			WithOutputHandler(outputHandler.StdoutHandler, outputHandler.OutputChan)
		cmdPool.Start(runner)
	}
}

func ApplyDiskServerAttack(options core.AttackConfig, env Environment) error {
	var attackConf *core.DiskAttackConfig
	var ok bool
	if attackConf, ok = options.(*core.DiskAttackConfig); !ok {
		return fmt.Errorf("AttackConfig -> *DiskAttackConfig meet error")
	}
	poolSize := getPoolSize(attackConf)
	if attackConf.Action == core.DiskFillAction {
		cmdPool := pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool
		fillDisk(attackConf, cmdPool, NewOutputHandler(handleDiskServerOutput, nil))
		return nil
	}

	if attackConf.DdOptions != nil {
		var cmdPool *pkgUtils.CommandPools
		deadline := getDeadline(options)
		if deadline != nil {
			cmdPool = pkgUtils.NewCommandPools(context.Background(), deadline, poolSize)
		}
		cmdPool = pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool

		applyPayload(attackConf, cmdPool, NewOutputHandler(handleDiskServerOutput, nil))
	}
	return nil
}

func (diskServerAttack) Recover(exp core.Experiment, env Environment) error {
	attackConfig, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	config := *attackConfig.(*core.DiskAttackConfig)

	switch config.Action {
	case core.DiskFillAction, core.DiskWritePayloadAction:
		err = os.Remove(config.Path)
		if err != nil {
			log.Warn(fmt.Sprintf("recover disk: remove %s failed", config.Path), zap.Error(err))
		}
	}

	if cmdPool, ok := env.Chaos.CmdPools[exp.Uid]; ok {
		log.Info(fmt.Sprintf("stop disk attack,read: %s", config.Path))
		cmdPool.Close()
	}

	return nil
}
