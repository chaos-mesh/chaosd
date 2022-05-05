// Copyright 2022 Chaos Mesh Authors.
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

func NewVMAttackCommand(uid *string) *cobra.Command {
	options := core.NewVMOption()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.VMOption {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "vm attack",
		Short: "vm attack by using virsh",
		Run: func(*cobra.Command, []string) {
			options.Action = core.VMAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(vmAttack)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.VMName, "vm-name", "v", "", "The name of the vm to be destoryed")
	return cmd
}

func vmAttack(options *core.VMOption, chaos *chaosd.Server) {
	uid, err := chaos.ExecuteAttack(chaosd.VMAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("VM attack %v successfully, uid: %s", options, uid))
}
