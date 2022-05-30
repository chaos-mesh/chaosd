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

func NewKafkaAttackCommand(uid *string) *cobra.Command {
	options := core.NewKafkaCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.KafkaCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "kafka <subcommand>",
		Short: "Kafka attack related commands",
	}

	cmd.AddCommand(
		NewKafkaFillCommand(dep, options),
		NewKafkaFloodCommand(dep, options),
		NewKafkaIOCommand(dep, options),
	)

	cmd.PersistentFlags().StringVarP(&options.Topic, "topic", "T", "", "the topic to attack")
	_ = cmd.MarkPersistentFlagRequired("topic")

	return cmd
}

func NewKafkaFillCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "fill [options]",
		Short:             "Fill kafka cluster with messages",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaFillAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Host, "host", "H", "localhost", "the host of kafka server")
	cmd.Flags().Uint16VarP(&options.Port, "port", "P", 9092, "the port of kafka server")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "the username of kafka client")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "the password of kafka client")
	cmd.Flags().UintVarP(&options.MessageSize, "size", "s", 4*1024, "the size of each message")
	cmd.Flags().Uint64VarP(&options.MaxBytes, "max-bytes", "m", 1<<34, "the max bytes to fill")
	return cmd
}

func NewKafkaFloodCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "flood [options]",
		Short:             "Flood kafka cluster with messages",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaFloodAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}
	cmd.Flags().StringVarP(&options.Host, "host", "H", "localhost", "the host of kafka server")
	cmd.Flags().Uint16VarP(&options.Port, "port", "P", 9092, "the port of kafka server")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "the username of kafka client")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "the password of kafka client")
	cmd.Flags().UintVarP(&options.MessageSize, "size", "s", 1024, "the size of each message")
	cmd.Flags().UintVarP(&options.Threads, "threads", "t", 100, "the numbers of worker threads")
	return cmd
}

func NewKafkaIOCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "io [options]",
		Short:             "Make kafka cluster non-writable/non-readable",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaIOAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.ConfigFile, "config", "c", "/etc/kafka/server.properties", "the path of server config")
	cmd.Flags().BoolVarP(&options.NonReadable, "non-readable", "r", false, "make kafka cluster non-readable")
	cmd.Flags().BoolVarP(&options.NonWritable, "non-writable", "w", false, "make kafka cluster non-writable")
	return cmd
}

func kafkaCommandFunc(options *core.KafkaCommand, chaos *chaosd.Server) {
	options.CompleteDefaults()

	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	uid, err := chaos.ExecuteAttack(chaosd.KafkaAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack kafka successfully, uid: %s", uid))
}
