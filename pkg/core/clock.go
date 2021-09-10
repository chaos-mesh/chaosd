package core

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/time/utils"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"strings"
)

type ClockOption struct {
	CommonAttackConfig

	Pid           int
	SecDelta      int64
	NsecDelta     int64
	ClockIdsSlice string

	ClockIdsMask uint64
}

func NewClockOption() *ClockOption {
	return &ClockOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: ClockAttack,
		},
	}
}

func (opt *ClockOption) Preprocess() error {
	clkIds := strings.Split(opt.ClockIdsSlice, ",")

	clockIdsMask, err := utils.EncodeClkIds(clkIds)
	if err != nil {
		log.Error("error while converting clock ids to mask", zap.Error(err))
		return err
	}

	opt.ClockIdsMask = clockIdsMask
	return nil
}
