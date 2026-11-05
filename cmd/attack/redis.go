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
		Short: "Inject faults into Redis",
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
		Short: "Restart Redis Sentinel",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisSentinelRestartAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "Redis Sentinel address in host:port format")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "Redis Sentinel password")
	cmd.Flags().StringVarP(&options.Conf, "conf", "c", "", "Path to the Sentinel configuration file")
	cmd.Flags().BoolVarP(&options.FlushConfig, "flush-config", "", true, "Rewrite the Sentinel configuration file before shutdown")
	cmd.Flags().StringVarP(&options.RedisPath, "redis-path", "", "", "Directory containing redis-server, or empty to use PATH")

	return cmd
}

func NewRedisSentinelStopCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sentinel-stop",
		Short: "Stop Redis Sentinel",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisSentinelStopAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "Redis Sentinel address in host:port format")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "Redis Sentinel password")
	cmd.Flags().StringVarP(&options.Conf, "conf", "c", "", "Path to the Sentinel configuration file used for recovery")
	cmd.Flags().BoolVarP(&options.FlushConfig, "flush-config", "", true, "Rewrite the Sentinel configuration file before shutdown")
	cmd.Flags().StringVarP(&options.RedisPath, "redis-path", "", "", "Directory containing redis-server, or empty to use PATH")

	return cmd
}

func NewRedisCachePenetrationCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-penetration",
		Short: "Send repeated GET requests to a fixed Redis key",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCachePenetrationAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "Redis server address in host:port format")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "Redis server password")
	cmd.Flags().IntVarP(&options.RequestNum, "request-num", "", 0, "Number of GET requests to send")

	return cmd
}

func NewRedisCacheLimitCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-limit",
		Short: "Set the Redis maxmemory limit",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCacheLimitAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "Redis server address in host:port format")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "Redis server password")
	cmd.Flags().StringVarP(&options.CacheSize, "size", "s", "0", "Redis maxmemory value, such as 1048576 or 1mb (0 removes the limit)")
	cmd.Flags().StringVarP(&options.Percent, "percent", "", "", "Percentage of the current maxmemory limit, such as 50%")

	return cmd
}

func NewRedisCacheExpirationCommand(dep fx.Option, options *core.RedisCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache-expiration",
		Short: "Set expiration for Redis keys",
		Run: func(*cobra.Command, []string) {
			options.Action = core.RedisCacheExpirationAction
			utils.FxNewAppWithoutLog(dep, fx.Invoke(redisAttackF)).Run()
		},
	}

	cmd.Flags().StringVarP(&options.Addr, "addr", "a", "", "Redis server address in host:port format")
	cmd.Flags().StringVarP(&options.Password, "password", "p", "", "Redis server password")
	cmd.Flags().StringVarP(&options.Key, "key", "k", "", "Exact key to expire, or empty for all keys in the default database")
	cmd.Flags().StringVarP(&options.Expiration, "expiration", "", "0", "Expiration delay, such as 5s or 30m (0 requests immediate expiration)")
	cmd.Flags().StringVarP(&options.Option, "option", "", "", "Expiration condition: NX, XX, GT, or LT")

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
