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

	"github.com/hashicorp/go-multierror"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	pkgUtils "github.com/chaos-mesh/chaosd/pkg/utils"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type diskAttack struct{}

var DiskAttack AttackType = diskAttack{}

func handleDiskAttackOutput(output []byte, err error, c chan interface{}) {
	if err != nil {
		log.Error(string(output), zap.Error(err))
		c <- err
	}
	log.Info(string(output))
}

func (diskAttack) Attack(options core.AttackConfig, env Environment) error {
	err := ApplyDiskAttack(options, env)
	if err != nil {
		return err
	}
	return nil
}

func handleOutputChannelError(c chan interface{}) error {
	close(c)
	var multiErrs error
	for i := range c {
		if err, ok := i.(error); ok {
			multiErrs = multierror.Append(multiErrs, err)
		}
	}
	if multiErrs != nil {
		return multiErrs
	}
	return nil
}

func ApplyDiskAttack(options core.AttackConfig, env Environment) error {
	var attackConf *core.DiskAttackConfig
	var ok bool
	if attackConf, ok = options.(*core.DiskAttackConfig); !ok {
		return fmt.Errorf("AttackConfig -> *DiskAttackConfig meet error")
	}
	poolSize := getPoolSize(attackConf)
	outputChan := make(chan interface{}, poolSize+1)
	if attackConf.Action == core.DiskFillAction {
		cmdPool := pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool
		fillDisk(attackConf, cmdPool, NewOutputHandler(handleDiskAttackOutput, outputChan))
		cmdPool.Wait()
		cmdPool.Close()
		return handleOutputChannelError(outputChan)
	}

	if attackConf.DdOptions != nil {
		var cmdPool *pkgUtils.CommandPools
		deadline := getDeadline(options)
		if deadline != nil {
			cmdPool = pkgUtils.NewCommandPools(context.Background(), deadline, poolSize)
		}
		cmdPool = pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool

		applyPayload(attackConf, cmdPool, NewOutputHandler(handleDiskAttackOutput, outputChan))
		cmdPool.Wait()
		cmdPool.Close()
		return handleOutputChannelError(outputChan)
	}
	return nil
}

func (diskAttack) Recover(exp core.Experiment, env Environment) error {
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
		cmdPool.Close()
	}

	return nil
}
