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

package attack

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
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

	cmd.Flags().IntVarP(&options.Pid, "pid", "p", 0, "Pid of target program.")
	cmd.Flags().StringVarP(&options.TimeOffset, "time-offset", "t", "", "Specifies the length of time offset.")
	cmd.Flags().StringVarP(&options.ClockIdsSlice, "clock-ids-slice", "c", "CLOCK_REALTIME",
		"The identifier of the particular clock on which to act."+
			"More clock description in linux kernel can be found in man page of clock_getres, clock_gettime, clock_settime."+
			"Muti clock ids should be split with \",\"")
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
