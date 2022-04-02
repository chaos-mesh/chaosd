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

package chaosd

import (
	"os/exec"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/pingcap/errors"
	"go.uber.org/zap"

	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type redisAttack struct{}

var RedisAttack AttackType = redisAttack{}

const (
	STATUSOK = "OK"
)

func (redisAttack) Attack(options core.AttackConfig, _ Environment) error {
	attack := options.(*core.RedisCommand)

	cli := redis.NewClient(&redis.Options{
		Addr:     attack.Addr,
		Password: attack.Password,
		DB:       attack.DB,
	})
	_, err := cli.Ping(cli.Context()).Result()
	if err != nil {
		return errors.WithStack(err)
	}
	defer cli.Close()

	switch attack.Action {
	case core.RedisSentinelRestartAction:
	case core.RedisSentinelStopAction:
		// Because redis.Client doesn't have the func `FlushConfig()`, a redis.SentinelClient has to be created
		sentinelCli := redis.NewSentinelClient(&redis.Options{
			Addr: attack.Addr,
		})
		result, err := sentinelCli.FlushConfig(sentinelCli.Context()).Result()
		if err != nil {
			return errors.WithStack(err)
		}
		if result != STATUSOK {
			return errors.WithStack(err)
		}

		result, err = cli.Shutdown(cli.Context()).Result()
		if err != nil {
			return errors.WithStack(err)
		}
		if result != STATUSOK {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (redisAttack) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	pcmd := config.(*core.ProcessCommand)
	if pcmd.Signal != int(syscall.SIGSTOP) {
		if pcmd.RecoverCmd == "" {
			return core.ErrNonRecoverableAttack.New("only SIGSTOP process attack and process attack with the recover-cmd are supported to recover")
		}

		rcmd := exec.Command("bash", "-c", pcmd.RecoverCmd)
		if err := rcmd.Start(); err != nil {
			return errors.WithStack(err)
		}

		log.Info("Execute recover-cmd successfully", zap.String("recover-cmd", pcmd.RecoverCmd))

	} else {
		for _, pid := range pcmd.PIDs {
			if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
