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
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	client "github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type kafkaAttack struct{}

var KafkaAttack AttackType = kafkaAttack{}

func (j kafkaAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.KafkaCommand)
	switch attack.Action {
	case core.KafkaFillAction:
		return attackKafkaFill(attack, env)
	case core.KafkaFloodAction:
		return attackKafkaFlood(context.TODO(), attack, env)
	case core.KafkaIOAction:
		return attackKafkaIO(attack, env)
	default:
		return nil
	}
}

func (j kafkaAttack) Recover(exp core.Experiment, env Environment) error {
	attack := new(core.KafkaCommand)
	if err := json.Unmarshal([]byte(exp.RecoverCommand), attack); err != nil {
		return perr.Wrap(err, "unmarshal kafka command")
	}
	if attack.Action == core.KafkaIOAction {
		return recoverKafkaIO(attack, env)
	}
	return nil
}

func attackKafkaFill(attack *core.KafkaCommand, env Environment) (err error) {
	// TODO: make it configurable
	const messagePerRequest = 1024
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err := client.DialLeader(context.Background(), "tcp", endpoint, attack.Topic, 0)
	if err != nil {
		return perr.Wrapf(err, "dial kafka leader %s", endpoint)
	}
	defer conn.Close()
	msg := make([]byte, attack.MessageSize)
	msgList := make([]client.Message, 0, messagePerRequest)
	for i := 0; i < messagePerRequest; i++ {
		msgList = append(msgList, client.Message{Value: msg})
	}

	for {
		_, err := conn.WriteMessages(msgList...)
		if err != nil {
			return perr.Wrap(err, "write messages")
		}
	}
}

func attackKafkaFlood(ctx context.Context, attack *core.KafkaCommand, env Environment) (err error) {
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err := client.DialLeader(context.Background(), "tcp", endpoint, attack.Topic, 0)
	if err != nil {
		return perr.Wrapf(err, "dial kafka leader %s", endpoint)
	}
	defer conn.Close()
	msg := make([]byte, attack.MessageSize)
	timeout := time.Second / time.Duration(attack.RequestPerSecond)
	wg := new(sync.WaitGroup)
	for i := 0; i < int(attack.Threads); i++ {
		logger := log.With(zap.String("thread", fmt.Sprintf("thread-%d", i)))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					succeeded := 0
					failed := 0
					counter := uint64(0)
					for time.Now().Sub(start) < time.Second && counter < attack.RequestPerSecond {
						counter++
						err := conn.SetWriteDeadline(time.Now().Add(timeout))
						if err != nil {
							failed++
							break
						}
						_, err = conn.Write(msg)
						if err != nil {
							failed++
							break
						}
						succeeded++
					}
					logger.Info(fmt.Sprintf("succeeded: %d, failed: %d", succeeded, failed))
					if time.Now().Sub(start) < time.Second {
						time.Sleep(start.Add(time.Second).Sub(time.Now()))
					}
				}
			}

		}()
	}
	wg.Wait()
	return nil
}

func attackKafkaIO(attack *core.KafkaCommand, env Environment) (err error) {
	return nil
}

func recoverKafkaIO(attack *core.KafkaCommand, env Environment) (err error) {
	return nil
}
