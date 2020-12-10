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
	"context"
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
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
	uid := args[0]

	expStore := mustExpStoreFromCmd()
	exp, err := expStore.FindByUid(context.Background(), uid)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	if exp == nil {
		ExitWithMsg(ExitError, fmt.Sprintf("experiment %s not found", uid))
	}

	if exp.Status != core.Success {
		ExitWithMsg(ExitError, fmt.Sprintf("can not recover %s experiment", exp.Status))
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	switch exp.Kind {
	case chaosd.ProcessAttack:
		pcmd := &core.ProcessCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), pcmd); err != nil {
			ExitWithError(ExitError, err)
		}

		if pcmd.Signal != int(syscall.SIGSTOP) {
			ExitWithMsg(ExitError, fmt.Sprintf("process attack %s not support to recover", uid))
		}

		if err := chaos.RecoverProcessAttack(uid, pcmd); err != nil {
			ExitWithError(ExitError, errors.Errorf("Recover experiment %s failed, %s", uid, err.Error()))
		}
	case chaosd.NetworkAttack:
		ncmd := &core.NetworkCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), ncmd); err != nil {
			ExitWithError(ExitError, err)
		}

		if err := chaos.RecoverNetworkAttack(uid, ncmd); err != nil {
			ExitWithError(ExitError, errors.Errorf("Recover experiment %s failed, %s", uid, err.Error()))
		}
	case chaosd.StressAttack:
		if err := chaos.RecoverStressAttack(uid, exp.RecoverCommand); err != nil {
			ExitWithError(ExitError, err)
		}
	default:
		ExitWithMsg(ExitError, fmt.Sprintf("chaos experiment kind %s not found", exp.Kind))
	}

	NormalExit(fmt.Sprintf("Recover %s successfully", uid))
}
