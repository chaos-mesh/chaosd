package attack

import (
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
)

func NewClockAttackCommand(uid *string) *cobra.Command {
	options := core.NewClockOption()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.ClockOption {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "clock attack",
		Short: "clock skew",
		Run: func(*cobra.Command, []string) {
			options.Action = "Attack"
			utils.FxNewAppWithoutLog(dep, fx.Invoke(processClockAttack)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.Pid, "pid", "p", 0, "")
	cmd.Flags().Int64VarP(&options.SecDelta, "second-delta", "s", 0, "")
	cmd.Flags().Int64VarP(&options.NsecDelta, "nanosecond-delta", "n", 0, "")
	cmd.Flags().StringVarP(&options.ClockIdsSlice, "clock-ids-slice", "c", "", "")

	return cmd
}

func processClockAttack(options *core.ClockOption, chaos *chaosd.Server) {

}
