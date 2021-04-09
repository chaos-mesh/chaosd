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
	"strconv"
)

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

type DiskCommand struct {
	Action          string `json:"action"`
	Size            string `json:"size"`
	Path            string `json:"path"`
	Percent         string `json:"percent"`
	FillByFallocate bool   `json:"fill_by_fallocate"`
}

func (d *DiskCommand) Validate() error {
	if d.Percent == "" && d.Size == "" {
		return fmt.Errorf("one of percent and size must not be empty, DiskCommand : %v", d)
	}
	if d.FillByFallocate && (d.Size == "0" || (d.Size == "" && d.Percent == "0")) {
		return fmt.Errorf("fallocate not suppurt 0 size or 0 percent data, "+
			"if you want allocate a 0 size file please set fallocate=false, DiskCommand : %v", d)
	}
	if _, err := strconv.ParseUint(d.Size, 10, 0); err != nil {
		return fmt.Errorf("unsupport size : %s, DiskCommand : %v", d.Size, d)
	}
	_, err := strconv.ParseUint(d.Percent, 10, 0)
	if d.Size == "" && err != nil {
		return fmt.Errorf("unsupport percent : %s, DiskCommand : %v", d.Percent, d)
	}
	return nil
}

func (d *DiskCommand) String() string {
	data, _ := json.Marshal(d)

	return string(data)
}
