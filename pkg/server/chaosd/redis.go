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

package chaosd

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type redisAttack struct{}

var RedisAttack AttackType = redisAttack{}

const (
	STATUSOK = "OK"
)

func (redisAttack) Attack(options core.AttackConfig, env Environment) error {
	attack := options.(*core.RedisCommand)

	cli := redis.NewClient(&redis.Options{
		Addr:     attack.Addr,
		Password: attack.Password,
	})
	_, err := cli.Ping(cli.Context()).Result()
	if err != nil {
		return errors.WithStack(err)
	}

	switch attack.Action {
	case core.RedisSentinelRestartAction:
		err := env.Chaos.shutdownSentinelServer(attack, cli)
		if err != nil {
			return errors.WithStack(err)
		}
		return env.Chaos.recoverSentinelStop(attack)

	case core.RedisSentinelStopAction:
		return env.Chaos.shutdownSentinelServer(attack, cli)

	case core.RedisCachePenetrationAction:
		pipe := cli.Pipeline()
		for i := 0; i < attack.RequestNum; i++ {
			pipe.Get(cli.Context(), "CHAOS_MESH_nqE3BWm7khHv")
		}
		_, err := pipe.Exec(cli.Context())
		if err != redis.Nil {
			return errors.WithStack(err)
		}

	case core.RedisCacheLimitAction:
		// `maxmemory` is an interface listwith content similar to `[maxmemory 1024]`
		maxmemory, err := cli.ConfigGet(cli.Context(), "maxmemory").Result()
		if err != nil {
			return errors.WithStack(err)
		}
		// Get the value of maxmemory
		attack.OriginCacheSize = fmt.Sprint(maxmemory[1])

		var cacheSize string
		if attack.Percent != "" {
			percentage, err := strconv.ParseFloat(attack.Percent[0:len(attack.Percent)-1], 64)
			if err != nil {
				return errors.WithStack(err)
			}
			originCacheSize, err := strconv.ParseFloat(attack.OriginCacheSize, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			cacheSize = fmt.Sprint(int(math.Floor(originCacheSize / 100.0 * percentage)))
		} else {
			cacheSize = attack.CacheSize
		}

		result, err := cli.ConfigSet(cli.Context(), "maxmemory", cacheSize).Result()
		if err != nil {
			return errors.WithStack(err)
		}
		if result != STATUSOK {
			return errors.WithStack(errors.Errorf("redis command status is %s", result))
		}
	}
	return nil
}

func (redisAttack) Recover(exp core.Experiment, env Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	attack := config.(*core.RedisCommand)

	cli := redis.NewClient(&redis.Options{
		Addr:     attack.Addr,
		Password: attack.Password,
	})

	switch attack.Action {
	case core.RedisSentinelStopAction:
		return env.Chaos.recoverSentinelStop(attack)

	case core.RedisCacheLimitAction:
		result, err := cli.ConfigSet(cli.Context(), "maxmemory", attack.OriginCacheSize).Result()
		if err != nil {
			return errors.WithStack(err)
		}
		if result != STATUSOK {
			return errors.WithStack(errors.Errorf("redis command status is %s", result))
		}
	}
	return nil
}

func (s *Server) shutdownSentinelServer(attack *core.RedisCommand, cli *redis.Client) error {
	if attack.FlushConfig {
		// Because redis.Client doesn't have the func `FlushConfig()`, a redis.SentinelClient has to be created
		sentinelCli := redis.NewSentinelClient(&redis.Options{
			Addr: attack.Addr,
		})
		result, err := sentinelCli.FlushConfig(sentinelCli.Context()).Result()
		if err != nil {
			return errors.WithStack(err)
		}
		if result != STATUSOK {
			return errors.WithStack(errors.Errorf("redis command status is %s", result))
		}
	}

	// If cli.Shutdown() runs successfully, the result will be nil and the err will be "connection refused"
	result, err := cli.Shutdown(cli.Context()).Result()
	if result != "" {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverSentinelStop(attack *core.RedisCommand) error {
	if attack.Conf == "" {
		return errors.WithStack(errors.Errorf("redis config does not exist"))
	}
	var redisPath string
	if attack.RedisPath != "" {
		redisPath = attack.RedisPath + "/redis-server"
	} else {
		redisPath = "redis-server"
	}
	recoverCmd := exec.Command(redisPath, attack.Conf, "--sentinel")
	_, err := recoverCmd.CombinedOutput()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
