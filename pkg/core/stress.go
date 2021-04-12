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
	"time"

	"github.com/pingcap/errors"
)

const (
	StressCPUAction = "cpu"
	StressMemAction = "mem"
)

type StressCommand struct {
	Action string

	Load int

	Workers int

	Size string

	Options []string

	Duration time.Duration

	StressngPid int32
}

func (s *StressCommand) Validate() error {
	if len(s.Action) == 0 {
		return errors.New("action not provided")
	}

	return nil
}

func (s *StressCommand) String() string {
	data, _ := json.Marshal(s)

	return string(data)
}
