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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
	attack := options.(*core.DiskCommand)

	if options.String() == core.DiskFillAction {
		return disk.attackDiskFill(attack)
	}
	return disk.attackDiskPayload(attack)
}

func (diskAttack) attackDiskPayload(payload *core.DiskCommand) error {
	switch payload.Action {
	case core.DiskWritePayloadAction:
		if payload.Path == "" {
			payload.Path = "/dev/null"
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf(DDWritePayloadCommand, payload.Path, "1M", payload.Size))
		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Error(string(output), zap.Error(err))
		} else {
			log.Info(string(output))
		}
		return err
	case core.DiskReadPayloadAction:
		if payload.Path == "" {
			err := errors.Errorf("empty read payload path")
			log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
			return err
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf(DDReadPayloadCommand, payload.Path, "1M", payload.Size))
		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Error(string(output), zap.Error(err))
		} else {
			log.Info(string(output))
		}
		return err
	default:
		err := errors.Errorf("invalid payload action")
		log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
		return err
	}
}

const DDFillCommand = "dd if=/dev/zero of=%s bs=%s count=%s iflag=fullblock"
const DDFallocateCommand = "fallocate -l %sM %s"

func (diskAttack) attackDiskFill(fill *core.DiskCommand) error {
	if fill.Path == "" {
		tempFile, err := ioutil.TempFile("", "example")
		if err != nil {
			log.Error("unexpected err when open temp file", zap.Error(err))
			return err
		}

		if tempFile != nil {
			err = tempFile.Close()
			if err != nil {
				log.Error("unexpected err when close temp file", zap.Error(err))
				return err
			}
		} else {
			err := errors.Errorf("unexpected err : file get from ioutil.TempFile is nil")
			log.Error(fmt.Sprintf("payload action: %s", fill.Action), zap.Error(err))
			return err
		}

		fill.Path = tempFile.Name()
		defer func() {
			err := os.Remove(fill.Path)
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when removing temp file %s", fill.Path), zap.Error(err))
			}
		}()
	}

	var cmd *exec.Cmd
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
	}

	if fill.FillByFallocate {
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFallocateCommand, fill.Size, fill.Path))
	} else {
		//1M means the block size. The bytes size dd read | write is (block size) * (size).
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFillCommand, fill.Path, "1M", fill.Size))
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
