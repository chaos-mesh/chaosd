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

func NewJVMAttackCommand(uid *string) *cobra.Command {
	options := core.NewJVMCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.JVMCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "jvm <subcommand>",
		Short: "Inject faults into a JVM",
	}

	cmd.PersistentFlags().IntVarP(&options.Port, "port", "", 9288, "Byteman agent port")
	cmd.PersistentFlags().IntVarP(&options.Pid, "pid", "", 0, "PID of the Java process to attach to")

	cmd.AddCommand(
		NewJVMLatencyCommand(dep, options),
		NewJVMReturnCommand(dep, options),
		NewJVMExceptionCommand(dep, options),
		NewJVMStressCommand(dep, options),
		NewJVMGCCommand(dep, options),
		NewJVMRuleFileCommand(dep, options),
		NewJVMMySQLCommand(dep, options),
	)

	return cmd
}

func NewJVMLatencyCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latency [options]",
		Short: "Delay calls to a Java method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMLatencyAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "Java method name")
	cmd.Flags().IntVarP(&options.LatencyDuration, "latency", "", 0, "Delay in milliseconds")

	return cmd
}

func NewJVMReturnCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "return [options]",
		Short: "Override the return value of a Java method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMReturnAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "Java method name")
	cmd.Flags().StringVarP(&options.ReturnValue, "value", "", "", "Return value as a Byteman expression")

	return cmd
}

func NewJVMExceptionCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exception [options]",
		Short: "Throw an exception from a Java method",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMExceptionAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Class, "class", "c", "", "Java class name")
	cmd.Flags().StringVarP(&options.Method, "method", "m", "", "Java method name")
	cmd.Flags().StringVarP(&options.ThrowException, "exception", "", "", "Exception constructor with arguments, such as java.lang.RuntimeException(\"error\")")

	return cmd
}

func NewJVMStressCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress [options]",
		Short: "Stress JVM CPU or memory",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMStressAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().IntVarP(&options.CPUCount, "cpu-count", "", 0, "Number of CPU stress threads")
	cmd.Flags().StringVarP(&options.MemoryType, "mem-type", "", "", "Memory type to stress: stack or heap")

	return cmd
}

func NewJVMGCCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Trigger JVM garbage collection",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMGCAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	return cmd
}

func NewJVMRuleFileCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rule-file [options]",
		Short: "Inject faults from a Byteman rule file",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMRuleFileAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.RuleFile, "path", "p", "", "Path to the Byteman rule file")

	return cmd
}

func NewJVMMySQLCommand(dep fx.Option, options *core.JVMCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mysql [options]",
		Short: "Inject faults into a Java MySQL client",
		Run: func(*cobra.Command, []string) {
			options.Action = core.JVMMySQLAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(jvmCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.SQLType, "sql-type", "", "", "SQL statement type to match, or empty for all types")
	cmd.Flags().StringVarP(&options.Database, "database", "d", "", "Database name in SQL to match, or empty for all databases")
	cmd.Flags().StringVarP(&options.Table, "table", "t", "", "Table name to match, or empty for all tables")
	cmd.Flags().StringVarP(&options.MySQLConnectorVersion, "mysql-connector-version", "v", "8", "MySQL Connector/J major version: 5 or 8")
	cmd.Flags().StringVarP(&options.ThrowException, "exception", "", "", "Exception message to throw instead of adding latency")
	cmd.Flags().IntVarP(&options.LatencyDuration, "latency", "", 0, "Delay in milliseconds when --exception is empty")

	return cmd
}

func jvmCommandFunc(options *core.JVMCommand, chaos *chaosd.Server) {
	options.CompleteDefaults()

	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.JVMAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack jvm successfully, uid: %s", uid))
}
