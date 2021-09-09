package attack

import (
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func NewClockAttackCommand(uid *string) *cobra.Command {
	options := core.NewClockOption()
	//dep := fx.Options(
	//	server.Module,
	//	fx.Provide(func() *core.ClockOption {
	//		options.UID = *uid
	//		return options
	//	}),
	//)

	cmd := &cobra.Command{
		Use:   "clock attack",
		Short: "clock skew",
		Run: func(*cobra.Command, []string) {
			options.Action = "Attack"
			//utils.FxNewAppWithoutLog(dep, fx.Invoke(_)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.Pid, "pid", "p", 0, "")
	cmd.Flags().Int64VarP(&options.SecDelta, "second-delta", "s", 0, "")
	cmd.Flags().Int64VarP(&options.NsecDelta, "nanosecond-delta", "n", 0, "")
	cmd.Flags().StringVarP(&options.ClockIdsSlice, "clock-ids-slice", "c", "", "")

	return cmd
}
