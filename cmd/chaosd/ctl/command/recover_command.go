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

/*
func RecoverExp(exp core.ExperimentStore, uid string) error {
	exp, err := expStore.FindByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	if exp == nil {
		return fmt.Sprintf("experiment %s not found", uid)
	}

	if exp.Status != core.Success {
		return fmt.Sprintf("can not recover %s experiment", exp.Status)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	switch exp.Kind {
	case core.ProcessAttack:
		pcmd := &core.ProcessCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), pcmd); err != nil {
			return err
		}

		if pcmd.Signal != int(syscall.SIGSTOP) {
			return fmt.Sprintf("process attack %s not support to recover", uid))
		}

		if err := chaos.RecoverProcessAttack(uid, pcmd); err != nil {
			return errors.Errorf("Recover experiment %s failed, %s", uid, err.Error())
		}
	case core.NetworkAttack:
		ncmd := &core.NetworkCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), ncmd); err != nil {
			return err
		}

		if err := chaos.RecoverNetworkAttack(uid, ncmd); err != nil {
			return errors.Errorf("Recover experiment %s failed, %s", uid, err.Error())
		}
	case core.StressAttack:
		scmd := &core.StressCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), scmd); err != nil {
			return err
		}

		if err := chaos.RecoverStressAttack(uid, scmd); err != nil {
			return err
		}
	default:
		return fmt.Sprintf("chaos experiment kind %s not found", exp.Kind)
	}

	return nil
}
*/
