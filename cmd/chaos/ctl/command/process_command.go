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
	"syscall"

	"github.com/pingcap/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
)

type processFlag struct {
	process string
	single  int
}

const (
	killProcessAction = "kill"
	stopProcessAction = "stop"
)

func (f *processFlag) valid(action string) error {
	if len(f.process) == 0 {
		return errors.New("process not provided")
	}

	return nil
}

var pFlag *processFlag

func NewProcessAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process <subcommand>",
		Short: "Process attack related commands",
	}

	cmd.AddCommand(
		NewProcessKillCommand(),
		NewProcessStopCommand(),
	)

	return cmd
}

func NewProcessKillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill [options]",
		Short: "kill process, default signal 9",
		Run:   processKillCommandFunc,
	}

	cmd.Flags().StringVarP(&pFlag.process, "process", "p", "", "The process name or the process ID")
	cmd.Flags().IntVarP(&pFlag.single, "single", "s", 9, "The signal number to send")
	cmd.Flags().StringVarP(&conf.Runtime, "runtime", "r", "docker", "current container runtime")
	cmd.Flags().StringVarP(&conf.Platform, "platform", "f", "local", "platform to deploy, default: local, supported platform: local, kubernetes")

	return cmd
}

func NewProcessStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [options]",
		Short: "stop process, this action will stop the process with SIGSTOP",

		Run: processStopCommandFunc,
	}

	cmd.Flags().StringVarP(&pFlag.process, "process", "p", "", "The process name or the process ID")
	pFlag.single = int(syscall.SIGSTOP)

	cmd.Flags().StringVarP(&conf.Runtime, "runtime", "r", "docker", "current container runtime")
	cmd.Flags().StringVarP(&conf.Platform, "platform", "f", "local", "platform to deploy, default: local, supported platform: local, kubernetes")

	return cmd
}

func processKillCommandFunc(cmd *cobra.Command, args []string) {
	if err := pFlag.valid(killProcessAction); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	processAttackF(cmd, pFlag)
}

func processStopCommandFunc(cmd *cobra.Command, args []string) {
	if err := pFlag.valid(stopProcessAction); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	processAttackF(cmd, pFlag)
}

func processAttackF(cmd *cobra.Command, f *processFlag) {
	chaos := mustChaosdFromCmd(cmd, conf)

	uid, err := chaos.ProcessAttack(&core.ProcessCommand{
		Process: pFlag.process,
		Signal:  syscall.Signal(f.single),
	})
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack process %s successfully, uid: %s", f.process, uid))
}
