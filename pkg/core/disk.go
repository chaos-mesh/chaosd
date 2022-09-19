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

package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

var _ AttackConfig = &DiskAttackConfig{}

type DiskAttackConfig struct {
	CommonAttackConfig
	DdOptions       *[]DdOption
	FAllocateOption *FAllocateOption
	Path            string
}

func (d DiskAttackConfig) RecoverData() string {
	data, _ := json.Marshal(d)

	return string(data)
}

var DdCommand = utils.Command{Name: "dd"}

type DdOption struct {
	ReadPath  string `dd:"if"`
	WritePath string `dd:"of"`
	BlockSize string `dd:"bs"`
	Count     string `dd:"count"`
	Iflag     string `dd:"iflag"`
	Oflag     string `dd:"oflag"`
	Conv      string `dd:"conv"`
}

var FAllocateCommand = utils.Command{Name: "fallocate"}

type FAllocateOption struct {
	LengthOpt string `fallocate:"-"`
	Length    string `fallocate:"-"`
	FileName  string `fallocate:"-"`
}

type DiskOption struct {
	CommonAttackConfig

	Size              string `json:"size,omitempty"`
	Path              string `json:"path,omitempty"`
	Percent           string `json:"percent,omitempty"`
	PayloadProcessNum uint8  `json:"payload-process-num,omitempty"`

	FillByFallocate bool `json:"fallocate,omitempty"`
}

func NewDiskOption() *DiskOption {
	return &DiskOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: DiskAttack,
		},
		PayloadProcessNum: 1,
		FillByFallocate:   true,
	}
}

func (opt *DiskOption) PreProcess() (*DiskAttackConfig, error) {
	if err := opt.CommonAttackConfig.Validate(); err != nil {
		return nil, err
	}

	path, err := initPath(opt)
	if err != nil {
		return nil, err
	}

	byteSize, err := initSize(opt)
	if err != nil {
		return nil, err
	}

	if opt.Action == DiskFillAction && opt.FillByFallocate && byteSize != 0 {
		return &DiskAttackConfig{
			CommonAttackConfig: opt.CommonAttackConfig,
			DdOptions:          nil,
			FAllocateOption: &FAllocateOption{
				LengthOpt: "-l",
				Length:    strconv.FormatUint(byteSize, 10),
				FileName:  path,
			},
			Path: path,
		}, nil
	}

	ddOptions, err := initDdOptions(opt, path, byteSize)
	if err != nil {
		return nil, err
	}
	return &DiskAttackConfig{
		CommonAttackConfig: opt.CommonAttackConfig,
		DdOptions:          &ddOptions,
		FAllocateOption:    nil,
		Path:               path,
	}, nil
}

func initDdOptions(opt *DiskOption, path string, byteSize uint64) ([]DdOption, error) {
	ddBlocks, err := utils.SplitBytesByProcessNum(byteSize, opt.PayloadProcessNum)
	if err != nil {
		log.Error("fail to split disk size", zap.Error(err))
		return nil, err
	}
	var ddOpts []DdOption
	switch opt.Action {
	case DiskFillAction:
		for _, block := range ddBlocks {
			ddOpts = append(ddOpts, DdOption{
				ReadPath:  "/dev/zero",
				WritePath: path,
				BlockSize: block.BlockSize,
				Count:     block.Count,
				Iflag:     "fullblock", // fullblock : accumulate full blocks of input.
				Oflag:     "append",
				Conv:      "notrunc", // notrunc : do not truncate the output file.
			})
		}
	case DiskWritePayloadAction:
		for _, block := range ddBlocks {
			ddOpts = append(ddOpts, DdOption{
				ReadPath:  "/dev/zero",
				WritePath: path,
				BlockSize: block.BlockSize,
				Count:     block.Count,
				Oflag:     "dsync", // dsync : use synchronized I/O for data.
			})
		}
	case DiskReadPayloadAction:
		for _, block := range ddBlocks {
			ddOpts = append(ddOpts, DdOption{
				ReadPath:  path,
				WritePath: "/dev/null",
				BlockSize: block.BlockSize,
				Count:     block.Count,
				Iflag:     "dsync,fullblock,nocache", // nocache : Request to drop cache.
			})
		}
	}
	return ddOpts, nil
}

func initPath(opt *DiskOption) (string, error) {
	switch opt.Action {
	case DiskFillAction, DiskWritePayloadAction:
		if opt.Path == "" {
			var err error
			opt.Path, err = os.Getwd()
			if err != nil {
				log.Error("unexpected err when execute os.Getwd()", zap.Error(err))
				return "", err
			}
		}

		fi, err := os.Stat(opt.Path)
		if err != nil {
			// check if Path of file is valid when Path is not empty
			if os.IsNotExist(err) {
				var b []byte
				if err := ioutil.WriteFile(opt.Path, b, 0600); err != nil {
					return "", err
				}
				if err := os.Remove(opt.Path); err != nil {
					return "", err
				}
				return opt.Path, nil
			}
			return "", err
		}
		if fi.IsDir() {
			opt.Path, err = utils.CreateTempFile(opt.Path)
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err : %v , when CreateTempFile in action %s with path %s.", err, opt.Action, opt.Path))
				return "", err
			}
			if err := os.Remove(opt.Path); err != nil {
				return "", err
			}
			return opt.Path, err
		}
		return "", fmt.Errorf("fill into an existing file")
	case DiskReadPayloadAction:
		if opt.Path == "" {
			path, err := utils.GetRootDevice()
			if err != nil {
				log.Error("err when GetRootDevice in reading payload", zap.Error(err))
				return "", err
			}
			if path == "" {
				err = fmt.Errorf("can not get root device path")
				log.Error(fmt.Sprintf("payload action: %s", opt.Action), zap.Error(err))
				return "", err
			}
			return path, nil
		}
		var fi os.FileInfo
		var err error
		if fi, err = os.Stat(opt.Path); err != nil {
			return "", err
		}
		if fi.IsDir() {
			return "", fmt.Errorf("path is a dictory, path : %s", opt.Path)
		}
		f, err := os.Open(opt.Path)
		if err != nil {
			return "", err
		}
		err = f.Close()
		if err != nil {
			return "", nil
		}
		return opt.Path, nil
	default:
		return "", fmt.Errorf("unsupported action %s", opt.Action)
	}
}

func initSize(opt *DiskOption) (uint64, error) {
	if opt.Size != "" {
		byteSize, err := utils.ParseUnit(opt.Size)
		if err != nil {
			log.Error(fmt.Sprintf("fail to get parse size per units , %s", opt.Size), zap.Error(err))
			return 0, err
		}
		return byteSize, nil
	} else if opt.Percent != "" {
		opt.Percent = strings.Trim(opt.Percent, " %")
		percent, err := strconv.ParseUint(opt.Percent, 10, 0)
		if err != nil {
			log.Error(fmt.Sprintf("unexcepted err when parsing disk percent '%s'", opt.Percent), zap.Error(err))
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
	if opt.Action == DiskFillAction {
		return 0, fmt.Errorf("one of percent and size must not be empty, DiskOption : %v", opt)
	}
	return 0, fmt.Errorf("size must not be empty, DiskOption : %v", opt)
}
