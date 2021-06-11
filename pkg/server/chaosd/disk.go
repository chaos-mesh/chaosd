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
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type diskAttack struct{}

var DiskAttack AttackType = diskAttack{}

const DDWritePayloadCommand = "dd if=/dev/zero of=%s bs=%s count=%s oflag=dsync"
const DDReadPayloadCommand = "dd if=%s of=/dev/null bs=%s count=%s iflag=dsync,fullblock,nocache"

func (disk diskAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.DiskOption)

	if options.String() == core.DiskFillAction {
		return disk.diskFill(attack)
	}
	return disk.diskPayload(attack)
}

func initWritePayloadPath(payload *core.DiskOption) error {
	var err error
	payload.Path, err = utils.CreateTempFile()
	if err != nil {
		log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", payload.Action))
		return err
	}
	return nil
}

func initReadPayloadPath(payload *core.DiskOption) error {
	path, err := utils.GetRootDevice()
	if err != nil {
		log.Error("err when GetRootDevice in reading payload", zap.Error(err))
		return err
	}
	if path == "" {
		err = errors.Errorf("can not get root device path")
		log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
		return err
	}
	payload.Path = path
	return nil
}

// diskPayload will execute a dd command (DDWritePayloadCommand or DDReadPayloadCommand)
// to add a write or read payload.
func (diskAttack) diskPayload(payload *core.DiskOption) error {
	var cmdFormat string
	switch payload.Action {
	case core.DiskWritePayloadAction:
		cmdFormat = DDWritePayloadCommand
		if payload.Path == "" {
			err := initWritePayloadPath(payload)
			if err != nil {
				return err
			}
		}
	case core.DiskReadPayloadAction:
		cmdFormat = DDReadPayloadCommand
		if payload.Path == "" {
			err := initReadPayloadPath(payload)
			if err != nil {
				return err
			}
		}
	default:
		err := errors.Errorf("invalid payload action")
		log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
		return err
	}

	byteSize, err := utils.ParseUnit(payload.Size)
	if err != nil {
		log.Error(fmt.Sprintf("fail to get parse size per units , %s", payload.Size), zap.Error(err))
		return err
	}
	ddBlocks, err := utils.SplitBytesByProcessNum(byteSize, payload.PayloadProcessNum)
	if err != nil {
		log.Error(fmt.Sprintf("split size ,process num %d", payload.PayloadProcessNum), zap.Error(err))
		return err
	}
	if len(ddBlocks) == 0 {
		return nil
	}
	rest := ddBlocks[len(ddBlocks)-1]
	ddBlocks = ddBlocks[:len(ddBlocks)-1]
	cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdFormat, payload.Path, rest.BlockSize, rest.Count))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(cmd.String()+string(output), zap.Error(err))
	}
	log.Info(string(output))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error
	wg.Add(len(ddBlocks))
	for _, sizeBlock := range ddBlocks {
		cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdFormat, payload.Path, sizeBlock.BlockSize, sizeBlock.Count))

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

	return nil
}

// dd command with 'oflag=append conv=notrunc' will append new data in the file.
const DDFillCommand = "dd if=/dev/zero of=%s bs=%s count=%s iflag=fullblock oflag=append conv=notrunc"
const FallocateCommand = "fallocate -l %s %s"

// diskFill will execute a dd command (DDFillCommand or FallocateCommand)
// to fill the disk.
func (diskAttack) diskFill(fill *core.DiskOption) error {
	if fill.Path == "" {
		var err error
		fill.Path, err = utils.CreateTempFile()
		if err != nil {
			log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", fill.Action))
			return err
		}
	}

	if fill.Size != "" {
		fill.Size = strings.Trim(fill.Size, " ")
	} else if fill.Percent != "" {
		fill.Percent = strings.Trim(fill.Percent, " ")
		percent, err := strconv.ParseUint(fill.Percent, 10, 0)
		if err != nil {
			log.Error(fmt.Sprintf(" unexcepted err when parsing disk percent '%s'", fill.Percent), zap.Error(err))
			return err
		}
		dir := filepath.Dir(fill.Path)
		totalSize, err := utils.GetDiskTotalSize(dir)
		if err != nil {
			log.Error("fail to get disk total size", zap.Error(err))
			return err
		}
		fill.Size = strconv.FormatUint(totalSize*percent/100, 10) + "c"
	}
	var cmd *exec.Cmd
	if fill.FillByFallocate {
		cmd = exec.Command("bash", "-c", fmt.Sprintf(FallocateCommand, fill.Size, fill.Path))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(string(output), zap.Error(err))
			return err
		}
		log.Info(string(output))
	} else {
		byteSize, err := utils.ParseUnit(fill.Size)
		if err != nil {
			log.Error("fail to parse disk size", zap.Error(err))
			return err
		}

		ddBlocks, err := utils.SplitBytesByProcessNum(byteSize, 1)
		if err != nil {
			log.Error("fail to split disk size", zap.Error(err))
			return err
		}
		for _, block := range ddBlocks {
			cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFillCommand, fill.Path, block.BlockSize, block.Count))
			output, err := cmd.CombinedOutput()

			if err != nil {
				log.Error(string(output), zap.Error(err))
				return err
			}
			log.Info(string(output))
		}
	}

	return nil
}

func (diskAttack) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	option := *config.(*core.DiskOption)
	switch option.Action {
	case core.DiskFillAction, core.DiskWritePayloadAction:
		err = os.Remove(option.Path)
		if err != nil {
			log.Warn(fmt.Sprintf("recover disk: remove %s failed", option.Path), zap.Error(err))
		}
	}
	return nil
}
