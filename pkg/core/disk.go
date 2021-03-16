// Copyright 2020 Chaos Mesh Authors.
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

import "encoding/json"

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

type DiskCommand struct {
	Action string
	Size   uint64
	Path   string
}

func (d *DiskCommand) Validate() error {
	return nil
}

func (d *DiskCommand) String() string {
	data, _ := json.Marshal(d)

	return string(data)
}
