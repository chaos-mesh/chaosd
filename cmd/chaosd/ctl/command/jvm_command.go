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

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var jvmFlag core.JVMCommand

func NewJVMAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jvm <subcommand>",
		Short: "JVM attack related commands",
	}

	cmd.AddCommand(
		NewJVMPrepareCommand(),
		NewJVMAttachCommand(),
		//NewJVMAgentCommand(),
	)

	return cmd
}

func NewJVMPrepareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare [options]",
		Short: "attach agent to Java process for prepare",
		Run:   jvmPrepareCommandFunc,
	}

	cmd.Flags().IntVarP(&jvmFlag.Port, "port", "", 0, "")
	cmd.Flags().IntVarP(&jvmFlag.Pid, "pid", "", 0, "")
	jvmFlag.Type = core.JVMPrepareType

	return cmd
}

func NewJVMAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach [options]",
		Short: "attach Java process to inject fault",
		Run:   jvmAttachCommandFunc,
	}

	cmd.Flags().StringVarP(&jvmFlag.Name, "name", "n", "", "rule name, should be unique.")
	cmd.Flags().StringVarP(&jvmFlag.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&jvmFlag.Method, "method", "m", "", "the method name in Java class")
	cmd.Flags().StringVarP(&jvmFlag.Action, "action", "a", "", "fault action, values can be latency, exception, return")
	cmd.Flags().IntVarP(&jvmFlag.Port, "port", "", 0, "port")
	cmd.Flags().StringVarP(&jvmFlag.Value, "value", "v", "", "value")

	jvmFlag.Type = core.JVMAttachType

	return cmd
}

func jvmPrepareCommandFunc(cmd *cobra.Command, args []string) {
	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.JVMPrepare(&jvmFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack stress %s successfully, uid: %s", jvmFlag.Action, uid))
}

func jvmAttachCommandFunc(cmd *cobra.Command, args []string) {
	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.JVMAttack(&jvmFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack stress %s successfully, uid: %s", jvmFlag.Action, uid))
}
