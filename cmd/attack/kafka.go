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
		Short: "Inject faults into Kafka",
	}

	cmd.AddCommand(
		NewKafkaFillCommand(dep, options),
		NewKafkaFloodCommand(dep, options),
		NewKafkaIOCommand(dep, options),
	)

	cmd.PersistentFlags().StringVarP(&options.Topic, "topic", "T", "", "Kafka topic to attack")
	_ = cmd.MarkPersistentFlagRequired("topic")

	return cmd
}

func NewKafkaFillCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "fill [options]",
		Short:             "Fill a Kafka topic with messages",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaFillAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Host, "host", "H", "localhost", "Kafka broker host")
	cmd.Flags().Uint16VarP(&options.Port, "port", "P", 9092, "Kafka broker port")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "SASL username, or empty to disable authentication")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "SASL password")
	cmd.Flags().StringVarP(&options.AuthMechanism, "auth-mechanism", "a", "sasl/plain", "SASL mechanism: sasl/plain, sasl/scram-sha-256, or sasl/scram-sha-512")
	cmd.Flags().UintVarP(&options.MessageSize, "size", "s", 4*1024, "Message payload size in bytes")
	cmd.Flags().Uint64VarP(&options.MaxBytes, "max-bytes", "m", 1<<34, "Target total bytes to write")
	cmd.Flags().StringVarP(&options.ReloadCommand, "reload-cmd", "r", "", "Shell command to reload Kafka configuration")
	return cmd
}

func NewKafkaFloodCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "flood [options]",
		Short:             "Continuously send messages to a Kafka topic",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaFloodAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}
	cmd.Flags().StringVarP(&options.Host, "host", "H", "localhost", "Kafka broker host")
	cmd.Flags().Uint16VarP(&options.Port, "port", "P", 9092, "Kafka broker port")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "SASL username, or empty to disable authentication")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "SASL password")
	cmd.Flags().StringVarP(&options.AuthMechanism, "auth-mechanism", "a", "sasl/plain", "SASL mechanism: sasl/plain, sasl/scram-sha-256, or sasl/scram-sha-512")
	cmd.Flags().UintVarP(&options.MessageSize, "size", "s", 1024, "Message payload size in bytes")
	cmd.Flags().UintVarP(&options.Threads, "threads", "t", 100, "Number of concurrent message producers")
	return cmd
}

func NewKafkaIOCommand(dep fx.Option, options *core.KafkaCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "io [options]",
		Short:             "Remove read or write permissions from local Kafka topic files",
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(*cobra.Command, []string) {
			options.Action = core.KafkaIOAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(kafkaCommandFunc)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Host, "host", "H", "localhost", "Kafka broker host")
	cmd.Flags().Uint16VarP(&options.Port, "port", "P", 9092, "Kafka broker port")
	cmd.Flags().StringVarP(&options.Username, "username", "u", "", "SASL username, or empty to disable authentication")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "SASL password")
	cmd.Flags().StringVarP(&options.AuthMechanism, "auth-mechanism", "a", "sasl/plain", "SASL mechanism: sasl/plain, sasl/scram-sha-256, or sasl/scram-sha-512")
	cmd.Flags().StringVarP(&options.ConfigFile, "config", "c", "/etc/kafka/server.properties", "Path to the local Kafka server configuration file")
	cmd.Flags().BoolVarP(&options.NonReadable, "non-readable", "r", false, "Remove read permissions from local topic files")
	cmd.Flags().BoolVarP(&options.NonWritable, "non-writable", "w", false, "Remove write permissions from local topic files")
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
