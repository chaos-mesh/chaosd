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

func NewRedisAttackCommand(uid *string) *cobra.Command {
	options := core.NewRedisCommand()
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.RedisCommand {
			options.UID = *uid
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "redis <subcommand>",
		Short: "Redis attack related commands",
	}

	cmd.AddCommand(
		NewRedisSentinelRestartCommand(dep, options),
		NewRedisSentinelStopCommand(dep, options),
		NewRedisCachePenetrationCommand(dep, options),
		NewRedisCacheLimitCommand(dep, options),
		NewRedisCacheExpirationCommand(dep, options),
	)

	return cmd
}

func NewRedisSentinelRestartCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sentinel-restart",
		Short: "restart sentinel",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisSentinelRestartAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "The address of redis server")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "The password of server")
	cmd.Flags().StringVarP(&options.Conf, "conf", "c", "", "The config of Redis server")
	cmd.Flags().BoolVarP(&options.FlushConfig, "flush-config", "", true, " Force Sentinel to rewrite its configuration on disk")
	cmd.Flags().StringVarP(&options.RedisPath, "redis-path", "", "", "The path of the redis-server command")

	return cmd
}

func NewRedisSentinelStopCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sentinel-stop",
		Short: "stop sentinel",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisSentinelStopAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "The address of redis server")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "The password of server")
	cmd.Flags().StringVarP(&options.Conf, "conf", "c", "", "The config path of Redis server")
	cmd.Flags().BoolVarP(&options.FlushConfig, "flush-config", "", true, "Force Sentinel to rewrite its configuration on disk")
	cmd.Flags().StringVarP(&options.RedisPath, "redis-path", "", "", "The path of the redis-server command")

	return cmd
}

func NewRedisCachePenetrationCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-penetration",
		Short: "penetrate cache",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCachePenetrationAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "The address of redis server")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "The password of server")
	cmd.Flags().IntVarP(&options.RequestNum, "request-num", "", 0, "The number of requests")

	return cmd
}

func NewRedisCacheLimitCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-limit",
		Short: "set maxmemory of Redis",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCacheLimitAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "The address of redis server")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "The password of server")
	cmd.Flags().StringVarP(&options.CacheSize, "size", "s", "0", "The size of cache")
	cmd.Flags().StringVarP(&options.Percent, "percent", "", "", "The percentage of maxmemory")

	return cmd
}

func NewRedisCacheExpirationCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-expiration",
		Short: "expire keys in Redis",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCacheExpirationAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "The address of redis server")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "The password of server")
	cmd.Flags().StringVarP(&options.Key, "key", "k", "", "The key to be set a expiration, default expire all keys")
	cmd.Flags().StringVarP(&options.Expiration, "expiration", "", "0", `The expiration of the key. A expiration string should be able to be converted to a time duration, such as "5s" or "30m"`)
	cmd.Flags().StringVarP(&options.Option, "option", "", "", "The additional options of expiration, only NX, XX, GT, LT supported")

	return cmd
}

func redisAttackF(chaos *chaosd.Server, options *core.RedisCommand) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}
	uid, err := chaos.ExecuteAttack(chaosd.RedisAttack, options, core.CommandMode)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Attack redis %s successfully, uid: %s", options.Action, uid))
}
