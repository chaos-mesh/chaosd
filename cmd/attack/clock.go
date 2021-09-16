package attack

import (
	"fmt"
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

	cmd.Flags().BoolVarP(&options.CheckPidExist, "check-pid-exist", "l", true, "")
	return cmd
}

func processClockAttack(options *core.ClockOption, chaos *chaosd.Server) {
	err := options.PreProcess()
	if err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.ClockAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Clock attack %v successfully, uid: %s", options, uid))
}
