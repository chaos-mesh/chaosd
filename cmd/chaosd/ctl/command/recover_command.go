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

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

func NewRecoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover UID",
		Short: "Recover a chaos experiment",
		Args:  cobra.MinimumNArgs(1),
		Run:   recoverCommandF,
	}

	cmd.Flags().StringVarP(&conf.Runtime, "runtime", "r", "docker", "current container runtime")
	cmd.Flags().StringVarP(&conf.Platform, "platform", "f", "local", "platform to deploy, default: local, supported platform: local, kubernetes")

	return cmd
}

func recoverCommandF(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		ExitWithMsg(ExitBadArgs, "UID is required")
	}
	uid := args[0]

	expStore := mustExpStoreFromCmd()
	chaos := mustChaosdFromCmd(cmd, &conf)

	err := utils.RecoverExp(expStore, chaos, uid)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Recover %s successfully", uid))
}
