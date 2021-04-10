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
)

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

type DiskCommand struct {
	CommonAttackConfig

	Size            uint64 `json:"size"`
	Path            string `json:"path"`
	FillByFallocate bool   `json:"fill_by_fallocate"`
}

var _ AttackConfig = &DiskCommand{}

func (d DiskCommand) Validate() error {
	if d.Action == DiskFillAction || d.Action == DiskWritePayloadAction || d.Action == DiskReadPayloadAction {
		return nil
	}
	return fmt.Errorf("invalid disk attack action %v", d.Action)
}

func (d DiskCommand) RecoverData() string {
	data, _ := json.Marshal(d)

	return string(data)
}

func NewDiskCommand() *DiskCommand {
	return &DiskCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: DiskAttack,
		},
	}
}
