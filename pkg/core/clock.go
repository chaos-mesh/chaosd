package core

type ClockOption struct {
	CommonAttackConfig

	Pid           int
	SecDelta      int64
	NsecDelta     int64
	ClockIdsSlice string
}

func NewClockOption() *ClockOption {
	return &ClockOption{
		CommonAttackConfig: CommonAttackConfig{
			Kind: ClockAttack,
		},
	}
}
