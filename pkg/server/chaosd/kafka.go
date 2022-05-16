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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/magiconair/properties"
	"github.com/pingcap/log"
	perr "github.com/pkg/errors"
	client "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type kafkaAttack struct{}

var KafkaAttack AttackType = kafkaAttack{}

func (j kafkaAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.KafkaCommand)
	switch attack.Action {
	case core.KafkaFillAction:
		return attackKafkaFill(attack)
	case core.KafkaFloodAction:
		return attackKafkaFlood(context.TODO(), attack)
	case core.KafkaIOAction:
		return attackKafkaIO(attack)
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

func dial(attack *core.KafkaCommand) (conn *client.Conn, err error) {
	dialer := &client.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if attack.Username != "" {
		dialer.SASLMechanism, err = scram.Mechanism(scram.SHA512, attack.Username, attack.Password)
		if err != nil {
			return nil, perr.Wrap(err, "create scram mechanism")
		}
	}
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err = dialer.DialLeader(context.Background(), "tcp", endpoint, attack.Topic, attack.Partition)
	if err != nil {
		return nil, perr.Wrapf(err, "dial kafka leader %s", endpoint)
	}
	return conn, nil
}

func attackKafkaFill(attack *core.KafkaCommand) (err error) {
	// TODO: make it configurable
	const messagePerRequest = 128
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err := client.DialLeader(context.Background(), "tcp", endpoint, attack.Topic, attack.Partition)
	if err != nil {
		return perr.Wrapf(err, "dial kafka leader %s", endpoint)
	}
	defer conn.Close()
	msg := make([]byte, attack.MessageSize)
	msgList := make([]client.Message, 0, messagePerRequest)
	for i := 0; i < messagePerRequest; i++ {
		msgList = append(msgList, client.Message{Value: msg})
	}

	counter := uint64(0)
	start := time.Now()
	written := "0 B"

	for counter < attack.MaxBytes {
		n, err := conn.WriteMessages(msgList...)
		if err != nil {
			return perr.Wrap(err, "write messages")
		}
		counter += uint64(n)
		newWritten := humanize.Bytes(counter)
		if newWritten != written {
			written = newWritten
			log.Info(fmt.Sprintf("write %s in %s", written, time.Now().Sub(start)))
		}
	}
	return nil
}

func attackKafkaFlood(ctx context.Context, attack *core.KafkaCommand) (err error) {
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err := client.DialLeader(context.Background(), "tcp", endpoint, attack.Topic, attack.Partition)
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
							if attack.NoSilent {
								logger.Error("set write deadline", zap.Error(err))
							}
							continue
						}
						_, err = conn.Write(msg)
						if err != nil {
							failed++
							if attack.NoSilent {
								logger.Error("write message", zap.Error(err))
							}
							continue
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

func attackKafkaIO(attack *core.KafkaCommand) error {
	p, err := properties.LoadFile(attack.ConfigFile, properties.UTF8)
	if err != nil {
		return perr.Wrapf(err, "load config file %s", attack.ConfigFile)
	}
	attack.PartitionDir, err = findPartitionDir(attack, strings.Split(p.GetString("log.dirs", "/var/lib/kafka"), ","))
	if err != nil {
		return err
	}

	dir, err := os.Stat(attack.PartitionDir)
	if err != nil {
		return perr.Wrapf(err, "stat partition dir %s", attack.PartitionDir)
	}
	mode := dir.Mode()
	if attack.NonReadable {
		mode &= ^os.FileMode(0444)
	}
	if attack.NonWritable {
		mode &= ^os.FileMode(0200)
	}
	if attack.NoSilent {
		log.Info(fmt.Sprintf("change permission of %s to %s", attack.PartitionDir, mode))
	}
	err = os.Chmod(attack.PartitionDir, mode)
	if err != nil {
		return perr.Wrapf(err, "change permission of %s", dir.Name())
	}

	files, err := os.ReadDir(attack.PartitionDir)
	if err != nil {
		return perr.Wrapf(err, "read partition dir %s", attack.PartitionDir)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".log") {
			continue
		}
		filePath := path.Join(attack.PartitionDir, file.Name())
		f, err := os.Stat(filePath)
		if err != nil {
			return perr.Wrapf(err, "stat file %s", filePath)
		}

		mode := f.Mode()
		if attack.NonReadable {
			mode &= ^os.FileMode(0444)
		}
		if attack.NonWritable {
			mode &= ^os.FileMode(0200)
		}
		if attack.NoSilent {
			log.Info(fmt.Sprintf("change permission of %s to %s", filePath, mode))
		}
		err = os.Chmod(filePath, mode)
		if err != nil {
			return perr.Wrapf(err, "change permission of %s", file.Name())
		}
	}
	return nil
}

func recoverKafkaIO(attack *core.KafkaCommand, env Environment) (err error) {
	dir, err := os.Stat(attack.PartitionDir)
	if err != nil {
		return perr.Wrapf(err, "stat partition dir %s", attack.PartitionDir)
	}
	mode := dir.Mode()
	if attack.NonReadable {
		mode |= os.FileMode(0444)
	}
	if attack.NonWritable {
		mode |= os.FileMode(0200)
	}
	if attack.NoSilent {
		log.Info(fmt.Sprintf("change permission of %s to %s", attack.PartitionDir, mode))
	}
	err = os.Chmod(attack.PartitionDir, mode)
	if err != nil {
		return perr.Wrapf(err, "change permission of %s", dir.Name())
	}

	files, err := os.ReadDir(attack.PartitionDir)
	if err != nil {
		return perr.Wrapf(err, "read partition dir %s", attack.PartitionDir)
	}
	for _, file := range files {
		filePath := path.Join(attack.PartitionDir, file.Name())
		f, err := os.Stat(filePath)
		if err != nil {
			return perr.Wrapf(err, "stat file %s", filePath)
		}
		mode := f.Mode()
		if attack.NonReadable {
			mode |= os.FileMode(0444)
		}
		if attack.NonWritable {
			mode |= os.FileMode(0200)
		}
		if attack.NoSilent {
			log.Info(fmt.Sprintf("change permission of %s to %s", filePath, mode))
		}
		err = os.Chmod(filePath, mode)
		if err != nil {
			return perr.Wrapf(err, "change permission of %s", attack.PartitionDir)
		}
	}
	return nil
}

func findPartitionDir(attack *core.KafkaCommand, logDirs []string) (string, error) {
	dirName := fmt.Sprintf("%s-%d", attack.Topic, attack.Partition)
	for _, dir := range logDirs {
		entries, err := os.ReadDir(strings.TrimSpace(dir))
		if err != nil {
			if attack.NoSilent {
				log.Error("read dir", zap.Error(err))
			}
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == dirName {
				return path.Join(strings.TrimSpace(dir), entry.Name()), nil
			}
		}

	}
	return "", perr.Errorf("partition dir %s not found", dirName)
}
