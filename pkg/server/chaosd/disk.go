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

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type diskAttack struct{}

var DiskAttack AttackType = diskAttack{}

const DDWritePayloadCommand = "dd if=/dev/zero of=%s bs=%s count=%s oflag=dsync"
const DDReadPayloadCommand = "dd if=%s of=/dev/null bs=%s count=%s iflag=dsync,direct,fullblock"

func (disk diskAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.DiskOption)

	if options.String() == core.DiskFillAction {
		return disk.attackDiskFill(attack)
	}
	return disk.attackDiskPayload(attack)
}

func (diskAttack) attackDiskPayload(payload *core.DiskOption) error {
	var cmdFormat string
	switch payload.Action {
	case core.DiskWritePayloadAction:
		cmdFormat = DDWritePayloadCommand
		if payload.Path == "" {
			var err error
			payload.Path, err = utils.CreateTempFile()
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", payload.Action))
				return err
			}
			defer func() {
				err := os.Remove(payload.Path)
				if err != nil {
					log.Error(fmt.Sprintf("unexpected err when removing temp file %s", payload.Path), zap.Error(err))
				}
			}()
		}
	case core.DiskReadPayloadAction:
		cmdFormat = DDReadPayloadCommand
		if payload.Path == "" {
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
		}
	default:
		err := errors.Errorf("invalid payload action")
		log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
		return err
	}

	byteSize, err := utils.ParseUnit(payload.Size + payload.Unit)
	if err != nil {
		log.Error(fmt.Sprintf("fail to get parse size unit , %s, %s", payload.Size, payload.Unit), zap.Error(err))
		return err
	}
	sizeBlocks := utils.SplitByteSize(byteSize, payload.PayloadProcessNum)

	var wg sync.WaitGroup
	wg.Add(len(sizeBlocks))
	fatalErrors := make(chan error)
	wgDone := make(chan bool)
	for _, sizeBlock := range sizeBlocks {
		cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdFormat, payload.Path, sizeBlock.BlockSize, sizeBlock.Size))

		go func(cmd *exec.Cmd) {
			output, err := cmd.CombinedOutput()

			if err != nil {
				log.Error(string(output), zap.Error(err))
				fatalErrors <- err
			}
			log.Info(string(output))
			wg.Done()
		}(cmd)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-fatalErrors:
		close(fatalErrors)
		return err
	}

	return nil
}

const DDFillCommand = "dd if=/dev/zero of=%s bs=%s count=%s iflag=fullblock"
const DDFallocateCommand = "fallocate -l %s%s %s"

func (diskAttack) attackDiskFill(fill *core.DiskOption) error {
	if fill.Path == "" {
		var err error
		fill.Path, err = utils.CreateTempFile()
		if err != nil {
			log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", fill.Action))
			return err
		}
	}

	defer func() {
		if fill.FillDestroyFile {
			err := os.Remove(fill.Path)
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when removing file %s", fill.Path), zap.Error(err))
			}
		}
	}()

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
		fill.Size = strconv.FormatUint(totalSize/1024/1024*percent/100, 10)
		fill.Unit = "M"
	}

	var cmd *exec.Cmd
	if fill.FillByFallocate {
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFallocateCommand, fill.Size, fill.Unit, fill.Path))
	} else {
		//1Unit means the block size. The bytes size dd read | write is (block size) * (size).
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFillCommand, fill.Path, fill.Unit, fill.Size))
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(string(output), zap.Error(err))
	} else {
		log.Info(string(output))
	}

	return err
}

func (diskAttack) Recover(exp core.Experiment, _ Environment) error {
	log.Info("Recover disk attack will do nothing, because delete | truncate data is too dangerous.")
	return nil
}
