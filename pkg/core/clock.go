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
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/time/utils"
)

type ClockOption struct {
	CommonAttackConfig

	Pid int

	TimeOffset string
	SecDelta   int64
	NsecDelta  int64

	ClockIdsSlice string

	Store ClockFuncStore

	ClockIdsMask uint64
}

type ClockFuncStore struct {
	CodeOfGetClockFunc []byte
	OriginAddress      uint64
}

func NewClockOption() *ClockOption {
	return &ClockOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: ClockAttack,
		},
	}
}

func (opt *ClockOption) PreProcess() error {
	clkIds := strings.Split(opt.ClockIdsSlice, ",")

	offset, err := time.ParseDuration(opt.TimeOffset)
	if err != nil {
		return err
	}
	opt.SecDelta = int64(offset / time.Second)
	opt.NsecDelta = offset.Nanoseconds()

	clockIdsMask, err := utils.EncodeClkIds(clkIds)
	if err != nil {
		log.Error("error while converting clock ids to mask", zap.Error(err))
		return err
	}
	if clockIdsMask == 0 {
		log.Error("clock ids must not be empty")
		return fmt.Errorf("clock ids must not be empty")
	}
	opt.ClockIdsMask = clockIdsMask

	if uint64(opt.SecDelta) > 1<<31 {
		log.Warn("Monotonic clock will be broken when sec delta is too large or too small.")
		if uint64(opt.SecDelta) > 1<<56 {
			log.Warn("Time zone info will be broken when sec delta is too large or too small.")
		}
	}

	if uint64(opt.NsecDelta) > 1<<56 {
		log.Warn("Time will be broken when nanosecond delta is too large or too small")
	}

	// Since os.FindProcess in unix systems will always succeed
	// regardless of whether the process exists (https://pkg.go.dev/os#FindProcess),
	// we need to use process.Signal to check if pid is accessible.
	process, err := os.FindProcess(opt.Pid)
	if err != nil {
		log.Error("failed to find process", zap.Error(err))
		return err
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			log.Error("pid may not be accessible", zap.Error(err))
			return err
		}
	}
	return nil
}

func (opt ClockOption) RecoverData() string {
	data, _ := json.Marshal(opt)

	return string(data)
}
