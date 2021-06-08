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

type DiskAttackConfig struct {
	core.CommonAttackConfig
	DdOptions        []DdOption
	FAllocateOptions []FAllocateOption
	Path             string
}

var ddCommand = utils.Command{Name: "dd"}

type DdOption struct {
	ReadPath  string `dd:"if"`
	WritePath string `dd:"of"`
	BlockSize string `dd:"bs"`
	Count     string `dd:"count"`
	Iflag     string `dd:"iflag"`
	Oflag     string `dd:"oflag"`
}

var fAllocateCommand = utils.Command{Name: "fallocate"}

type FAllocateOption struct {
	LengthOpt string `fallocate:"-"`
	Length    string `fallocate:"-"`
	FileName  string `fallocate:"-"`
}

func (cfg *DiskAttackConfig) Init(o *core.DiskOption) error {

}

func (cfg *DiskAttackConfig) path(opt *core.DiskOption) (string, error) {
	switch opt.Action {
	case DDFillCommand:
		if opt.Path == "" {
			var err error
			opt.Path, err = utils.CreateTempFile()
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", opt.Action))
				return "", err
			}
		} else {
			_, err := os.Stat(opt.Path)
			if err != nil {
				// check if Path of file is valid when Path is not empty
				if os.IsNotExist(err) {
					var b []byte
					if err := ioutil.WriteFile(opt.Path, b, 0644); err != nil {
						return "", err
					}
					if err := os.Remove(opt.Path); err != nil {
						return "", err
					}
				} else {
					return "", err
				}
			} else {
				return "", fmt.Errorf("fill into a existing file")
			}
		}
		return opt.Path, nil
	case DDReadPayloadCommand:
		path, err := utils.GetRootDevice()
		if err != nil {
			log.Error("err when GetRootDevice in reading payload", zap.Error(err))
			return "", err
		}
		if path == "" {
			err = errors.Errorf("can not get root device path")
			log.Error(fmt.Sprintf("payload action: %s", opt.Action), zap.Error(err))
			return "", err
		}
		return path, nil
	case DDWritePayloadCommand:
		path, err := utils.CreateTempFile()
		if err != nil {
			log.Error(fmt.Sprintf("unexpected err when CreateTempFile in action: %s", opt.Action))
			return "", err
		}
		return path, nil
	default:
		return "", errors.Errorf("unsupported action %s", opt.Action)
	}
}

func (cfg *DiskAttackConfig) size(opt *core.DiskOption) (uint64, error) {
	if opt.Size != "" {
		byteSize, err := utils.ParseUnit(opt.Size)
		if err != nil {
			log.Error(fmt.Sprintf("fail to get parse size per units , %s", opt.Size), zap.Error(err))
			return 0, err
		}
		return byteSize, nil
	} else if opt.Percent != "" {
		opt.Percent = strings.Trim(opt.Percent, " ")
		percent, err := strconv.ParseUint(opt.Percent, 10, 0)
		if err != nil {
			log.Error(fmt.Sprintf(" unexcepted err when parsing disk percent '%s'", opt.Percent), zap.Error(err))
			return 0, err
		}
		dir := filepath.Dir(opt.Path)
		totalSize, err := utils.GetDiskTotalSize(dir)
		if err != nil {
			log.Error("fail to get disk total size", zap.Error(err))
			return 0, err
		}
		return totalSize * percent / 100, nil
	}
	if opt.Action == DDFillCommand {
		return 0, fmt.Errorf("return fmt.Errorf(\"one of percent and size must not be empty, DiskOption : %v\", d)")
	} else {
		return 0, fmt.Errorf("return fmt.Errorf(\"one of percent and size must not be empty, DiskOption : %v\", d)")
	}

}

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
	if fill.FillByFAllocate {
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
