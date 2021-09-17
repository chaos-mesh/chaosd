package core

import (
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/time/utils"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"os"
	"strings"
	"syscall"
)

type ClockOption struct {
	CommonAttackConfig

	Pid           int
	SecDelta      int64
	NsecDelta     int64
	ClockIdsSlice string

	CheckPidExist bool
	Store         ClockFuncStore

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

	clockIdsMask, err := utils.EncodeClkIds(clkIds)
	if err != nil {
		log.Error("error while converting clock ids to mask", zap.Error(err))
		return err
	}
	opt.ClockIdsMask = clockIdsMask

	if opt.CheckPidExist {
		process, err := os.FindProcess(opt.Pid)
		if err != nil {
			fmt.Printf("Failed to find process: %s\n", err)
		} else {
			err := process.Signal(syscall.Signal(0))
			if err != nil {
				fmt.Printf("Pid may not be accessible , because : %v", err)
			}
		}
	}
	return nil
}

func (opt ClockOption) RecoverData() string {
	data, _ := json.Marshal(opt)

	return string(data)
}
