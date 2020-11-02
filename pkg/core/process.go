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

import (
	"encoding/json"
	"syscall"
)

type ProcessCommand struct {
	// Process defines the process name or the process ID.
	Process string
	Signal  syscall.Signal
	PIDs    []int
	// TODO: support these feature
	// Newest       bool
	// Oldest       bool
	// Exact        bool
	// KillChildren bool
	// User         string
}

func (p *ProcessCommand) String() string {
	data, _ := json.Marshal(p)

	return string(data)
}
