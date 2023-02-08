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

func handleFillingOutput(output []byte, err error) {
	if err != nil {
		log.Error(string(output), zap.Error(err))
	}
	log.Info(string(output))
}

func (disk diskServerAttack) Attack(options core.AttackConfig, env Environment) error {
	var attackConf *core.DiskAttackConfig
	var ok bool
	if attackConf, ok = options.(*core.DiskAttackConfig); !ok {
		return fmt.Errorf("AttackConfig -> *DiskAttackConfig meet error")
	}
	poolSize := 1
	if attackConf.DdOptions != nil && len(*attackConf.DdOptions) > 0 {
		poolSize = len(*attackConf.DdOptions)
	}
	if attackConf.Action == core.DiskFillAction {
		cmdPool := pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool

		if attackConf.FAllocateOption != nil {
			cmdPool.Start(core.FAllocateCommand, *attackConf.FAllocateOption, handleFillingOutput)
			return nil
		}

		for _, DdOption := range *attackConf.DdOptions {
			cmdPool.Start(core.DdCommand, DdOption, handleFillingOutput)
		}
		return nil
	}

	if attackConf.DdOptions != nil {
		duration, _ := options.ScheduleDuration()
		var cmdPool *pkgUtils.CommandPools
		if duration != nil {
			deadline := time.Now().Add(*duration)
			cmdPool = pkgUtils.NewCommandPools(context.Background(), &deadline, poolSize)
		}
		cmdPool = pkgUtils.NewCommandPools(context.Background(), nil, poolSize)
		env.Chaos.CmdPools[env.AttackUid] = cmdPool

		if len(*attackConf.DdOptions) == 0 {
			return nil
		}
		rest := (*attackConf.DdOptions)[len(*attackConf.DdOptions)-1]
		*attackConf.DdOptions = (*attackConf.DdOptions)[:len(*attackConf.DdOptions)-1]

		cmdPool.Start(core.DdCommand, rest, handleFillingOutput)

		for _, ddOpt := range *attackConf.DdOptions {
			cmdPool.Start(core.DdCommand, ddOpt, handleFillingOutput)
		}
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
		cmdPool.Close()
	}

	return nil
}
