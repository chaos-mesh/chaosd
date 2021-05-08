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

func NewJVMAttackCommand() *cobra.Command {
	options := core.NewJVMCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.JVMCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "jvm <subcommand>",
		Short: "JVM attack related commands",
	}

	cmd.AddCommand(
		NewJVMInstallCommand(dep, options),
		NewJVMSubmitCommand(dep, options),
	)

	return cmd
}

func NewJVMInstallCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [options]",
		Short: "install agent to Java process",
		Run: func(*cobra.Command, []string) {
			options.Type = core.JVMInstallType
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.Port, "port", "", 9288, "the port of agent server")
	cmd.Flags().IntVarP(&options.Pid, "pid", "", 0, "the pid of Java process which need to attach")

	return cmd
}

func NewJVMSubmitCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [options]",
		Short: "submit rules for byteman agent",
	}
	options.Type = core.JVMSubmitType

	cmd.PersistentFlags().StringVarP(&options.Name, "name", "n", "", "rule name, should be unique, and will generate one automatically if it is empty")
	cmd.PersistentFlags().IntVarP(&options.Port, "port", "", 9288, "the port of agent server")

	cmd.AddCommand(
		NewJVMLatencyCommand(dep, options),
		NewJVMReturnCommand(dep, options),
		NewJVMExceptionCommand(dep, options),
		NewJVMStressCommand(dep, options),
		NewJVMGCCommand(dep, options),
	)

	return cmd
}

func NewJVMLatencyCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latency [options]",
		Short: "inject latency to specified method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMLatencyAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "the method name in Java class")
	cmd.Flags().StringVarP(&options.LatencyDuration, "latency", "", "", "the latency duration")

	return cmd
}

func NewJVMReturnCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "return [options]",
		Short: "return specified value for specified method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMReturnAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "the method name in Java class")
	cmd.Flags().StringVarP(&options.ReturnValue, "value", "", "", "the return value for action 'return', only support number and string type now")

	return cmd
}

func NewJVMExceptionCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exception [options]",
		Short: "throw specified exception for specified method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMExceptionAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "the method name in Java class")
	cmd.Flags().StringVarP(&options.ThrowException, "exception", "", "", "the exception which needs to throw for action 'exception'")

	return cmd
}

func NewJVMStressCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress [options]",
		Short: "inject stress to JVM",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMStressAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.CPUCount, "cpu-count", "", 0, "the CPU core number need to use")
	cmd.Flags().IntVarP(&options.MemorySize, "mem-size", "", 0, "the memory size need to locate, the unit is MB")

	return cmd
}

func NewJVMGCCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gc",
		Short: "trigger GC for JVM",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMGCAction
			options.CompleteDefaults()
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	return cmd
}

func jvmCommandFunc(options *core.JVMCommand, chaos *chaosd.Server) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.JVMAttack, options)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack jvm successfully, uid: %s", uid))
}
