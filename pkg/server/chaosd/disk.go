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
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type diskAttack struct{}

var DiskAttack AttackType = diskAttack{}

func (disk diskAttack) Attack(options core.AttackConfig, env Environment) error {
	if attackConf, ok := options.(*core.DiskAttackConfig); ok {
		if attackConf.Action == core.DiskFillAction {
			if attackConf.FAllocateOption != nil {
				cmd := core.FAllocateCommand.Unmarshal(*attackConf.FAllocateOption)
				output, err := cmd.CombinedOutput()

				if err != nil {
					log.Error(string(output), zap.Error(err))
					return err
				}
				log.Info(string(output))
				return nil
			}

			for _, DdOption := range *attackConf.DdOptions {
				cmd := core.DdCommand.Unmarshal(DdOption)
				output, err := cmd.CombinedOutput()

				if err != nil {
					log.Error(string(output), zap.Error(err))
					return err
				}
				log.Info(string(output))
				return nil
			}
		}

		if attackConf.DdOptions != nil {
			if len(*attackConf.DdOptions) == 0 {
				return nil
			}
			rest := (*attackConf.DdOptions)[len(*attackConf.DdOptions)-1]
			*attackConf.DdOptions = (*attackConf.DdOptions)[:len(*attackConf.DdOptions)-1]

			cmd := core.DdCommand.Unmarshal(rest)
			output, err := cmd.CombinedOutput()

			if err != nil {
				log.Error(cmd.String()+string(output), zap.Error(err))
				return err
			}
			log.Info(string(output))

			var wg sync.WaitGroup
			var mu sync.Mutex
			var errs error
			wg.Add(len(*attackConf.DdOptions))
			for _, ddOpt := range *attackConf.DdOptions {
				cmd := core.DdCommand.Unmarshal(ddOpt)

				go func(cmd *exec.Cmd) {
					defer wg.Done()
					output, err := cmd.CombinedOutput()
					if err != nil {
						log.Error(cmd.String()+string(output), zap.Error(err))
						mu.Lock()
						defer mu.Unlock()
						errs = multierror.Append(errs, err)
						return
					}
					log.Info(string(output))
				}(cmd)
			}

			wg.Wait()

			if errs != nil {
				return errs
			}
		}
		return nil
	}
	return fmt.Errorf("AttackConfig -> *DiskAttackConfig meet error")
}

func (diskAttack) Recover(exp core.Experiment, _ Environment) error {
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
	return nil
}
