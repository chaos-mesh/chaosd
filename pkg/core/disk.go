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
	"strconv"

	"github.com/chaos-mesh/chaosd/pkg/utils"
)

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

type DiskOption struct {
	CommonAttackConfig

	Size              string `json:"size"`
	Path              string `json:"path"`
	Percent           string `json:"percent"`
	PayloadProcessNum uint8  `json:"payload_process_num"`

	FillByFallocate bool `json:"fill_by_fallocate"`
}

var _ AttackConfig = &DiskOption{}

func (d *DiskOption) Validate() error {
	if err := d.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	var byteSize uint64
	var err error
	if d.Size == "" {
		if d.Percent == "" {
			return fmt.Errorf("one of percent and size must not be empty, DiskOption : %v", d)
		}
		if byteSize, err = strconv.ParseUint(d.Percent, 10, 0); err != nil {
			return fmt.Errorf("unsupport percent : %s, DiskOption : %v", d.Percent, d)
		}
	} else {
		if byteSize, err = utils.ParseUnit(d.Size); err != nil {
			return fmt.Errorf("unknown units of size : %s, DiskOption : %v", d.Size, d)
		}
	}

	if d.Action == DiskFillAction || d.Action == DiskWritePayloadAction {
		if d.Action == DiskFillAction && d.FillByFallocate && byteSize == 0 {
			return fmt.Errorf("fallocate not suppurt 0 size or 0 percent data, "+
				"if you want allocate a 0 size file please set fallocate=false, DiskOption : %v", d)
		}

		_, err := os.Stat(d.Path)
		if err != nil {
			if os.IsNotExist(err) {
				// check if Path of file is valid when Path is not empty
				if d.Path != "" {
					var b []byte
					if err := ioutil.WriteFile(d.Path, b, 0644); err != nil {
						return err
					}
					if err := os.Remove(d.Path); err != nil {
						return err
					}
				}
			} else {
				return err
			}
		} else {
			if d.Action == DiskFillAction {
				return fmt.Errorf("fill into an existing file")
			}
			return fmt.Errorf("write into an existing file")
		}
	}

	if d.PayloadProcessNum == 0 {
		return fmt.Errorf("unsupport process num : %d, DiskOption : %v", d.PayloadProcessNum, d.Action)
	}

	return nil
}

func (d DiskOption) RecoverData() string {
	data, _ := json.Marshal(d)

	return string(data)
}

func NewDiskOption() *DiskOption {
	return &DiskOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: DiskAttack,
		},
	}
}
