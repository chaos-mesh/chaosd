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
	"github.com/pingcap/errors"
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
		return attackKafkaFill(context.TODO(), attack)
	case core.KafkaFloodAction:
		return attackKafkaFlood(context.TODO(), attack)
	case core.KafkaIOAction:
		return attackKafkaIO(attack)
	default:
		return errors.Errorf("invalid action: %s", attack.Action)
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

func newDialer(attack *core.KafkaCommand) (dialer *client.Dialer, err error) {
	dialer = &client.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if attack.Username != "" {
		dialer.SASLMechanism, err = scram.Mechanism(scram.SHA512, attack.Username, attack.Password)
		if err != nil {
			return nil, perr.Wrap(err, "create scram mechanism")
		}
	}
	return dialer, nil
}

func dial(attack *core.KafkaCommand) (conn *client.Conn, err error) {
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	dialer, err := newDialer(attack)
	if err != nil {
		return nil, perr.Wrap(err, "new dialer")
	}
	conn, err = dialer.Dial("tcp", endpoint)
	if err != nil {
		return nil, perr.Wrapf(err, "dial endpoint: %s", endpoint)
	}
	return conn, nil
}

func dialPartition(ctx context.Context, attack *core.KafkaCommand, partition int) (conn *client.Conn, err error) {
	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	dialer, err := newDialer(attack)
	if err != nil {
		return nil, perr.Wrap(err, "new dialer")
	}
	conn, err = dialer.DialLeader(ctx, "tcp", endpoint, attack.Topic, partition)
	if err != nil {
		return nil, perr.Wrapf(err, "dial endpoint: %s", endpoint)
	}
	return conn, nil
}

func getPartitions(attack *core.KafkaCommand) (partitions []int, err error) {
	dialer, err := newDialer(attack)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s:%d", attack.Host, attack.Port)
	conn, err := dialer.Dial("tcp", endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "dial endpoint: %s", endpoint)
	}

	if attack.Topic == "" {
		return nil, errors.New("topic is required")
	}
	pars, err := conn.ReadPartitions(attack.Topic)
	if err != nil {
		return nil, errors.Wrapf(err, "read partitions of topic %s", attack.Topic)
	}
	for _, par := range pars {
		if par.Error != nil {
			return nil, errors.Wrap(err, "read partition")
		}
		partitions = append(partitions, par.ID)
	}
	return partitions, nil
}

func attackKafkaFill(ctx context.Context, attack *core.KafkaCommand) (err error) {
	// TODO: make it configurable
	const messagePerRequest = 128
	msg := make([]byte, attack.MessageSize)
	msgList := make([]client.Message, 0, messagePerRequest)
	for i := 0; i < messagePerRequest; i++ {
		msgList = append(msgList, client.Message{Topic: attack.Topic, Value: msg})
	}

	partitions, err := getPartitions(attack)
	if err != nil {
		return err
	}

	writeChan := make(chan uint64)
	errChan := make(chan error)
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, partition := range partitions {
		p := partition
		go func() {
			conn, err := dialPartition(c, attack, p)
			if err != nil {
				errChan <- err
				return
			}

			for c.Err() == nil {
				n, err := conn.WriteMessages(msgList...)
				if err != nil {
					errChan <- perr.Wrap(err, "write messages")
					return
				}
				writeChan <- uint64(n)
			}
		}()
	}

	start := time.Now()
	written := "0 B"
	bytes := uint64(0)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			return err
		case n := <-writeChan:
			bytes += uint64(n)
			newWritten := humanize.Bytes(bytes)
			if newWritten != written {
				written = newWritten
				log.Info(fmt.Sprintf("write %s in %s", written, time.Now().Sub(start)))
			}
			if bytes >= attack.MaxBytes {
				return nil
			}
		}
	}
}

func attackKafkaFlood(ctx context.Context, attack *core.KafkaCommand) (err error) {
	msg := make([]byte, attack.MessageSize)
	wg := new(sync.WaitGroup)
	for i := 0; i < int(attack.Threads); i++ {
		logger := log.With(zap.String("thread", fmt.Sprintf("thread-%d", i)))
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := dialPartition(ctx, attack, 0)
			if err != nil {
				logger.Error(perr.Wrapf(err, "dial kafka broker: %s", fmt.Sprintf("%s:%d", attack.Host, attack.Port)).Error())
				return
			}
			defer conn.Close()
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
						conn.SetWriteDeadline(time.Now().Add(time.Second / time.Duration(attack.RequestPerSecond)))
						_, err = conn.WriteMessages(client.Message{Value: msg})
						if err != nil {
							failed++
							logger.Debug("write message", zap.Error(err))
							continue
						}
						succeeded++
						logger.Debug(fmt.Sprintf("time.Now().Sub(start) < time.Second: %t", time.Now().Sub(start) < time.Second))
						logger.Debug(fmt.Sprintf("counter < attack.RequestPerSecond: %t", counter < attack.RequestPerSecond))
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
	attack.PartitionDirs, err = findPartitionDirs(attack, strings.Split(p.GetString("log.dirs", "/var/lib/kafka"), ","))
	if err != nil {
		return err
	}

	for _, dirPath := range attack.PartitionDirs {
		dir, err := os.Stat(dirPath)
		if err != nil {
			return perr.Wrapf(err, "stat partition dir %s", dirPath)
		}
		mode := dir.Mode()
		if attack.NonReadable {
			mode &= ^os.FileMode(0444)
		}
		if attack.NonWritable {
			mode &= ^os.FileMode(0200)
		}
		log.Debug(fmt.Sprintf("change permission of %s to %s", dirPath, mode))
		err = os.Chmod(dirPath, mode)
		if err != nil {
			return perr.Wrapf(err, "change permission of %s", dirPath)
		}

		files, err := os.ReadDir(dirPath)
		if err != nil {
			return perr.Wrapf(err, "read partition dir %s", dirPath)
		}
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".log") {
				continue
			}
			filePath := path.Join(dirPath, file.Name())
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
			log.Debug(fmt.Sprintf("change permission of %s to %s", filePath, mode))
			err = os.Chmod(filePath, mode)
			if err != nil {
				return perr.Wrapf(err, "change permission of %s", file.Name())
			}
		}
	}

	return nil
}

func recoverKafkaIO(attack *core.KafkaCommand, env Environment) (err error) {
	for _, dirPath := range attack.PartitionDirs {
		dir, err := os.Stat(dirPath)
		if err != nil {
			return perr.Wrapf(err, "stat partition dir %s", dirPath)
		}
		mode := dir.Mode()
		if attack.NonReadable {
			mode |= os.FileMode(0444)
		}
		if attack.NonWritable {
			mode |= os.FileMode(0200)
		}
		log.S().Debugf("change permission of %s to %s", dirPath, mode)
		err = os.Chmod(dirPath, mode)
		if err != nil {
			return perr.Wrapf(err, "change permission of %s", dir.Name())
		}

		files, err := os.ReadDir(dirPath)
		if err != nil {
			return perr.Wrapf(err, "read partition dir %s", dirPath)
		}
		for _, file := range files {
			filePath := path.Join(dirPath, file.Name())
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
			log.S().Debugf("change permission of %s to %s", filePath, mode)
			err = os.Chmod(filePath, mode)
			if err != nil {
				return perr.Wrapf(err, "change permission of %s", dirPath)
			}
		}
	}
	return nil
}

func findPartitionDirs(attack *core.KafkaCommand, logDirs []string) ([]string, error) {
	partitions, err := getPartitions(attack)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(partitions))

	for _, partition := range partitions {
		dirName := fmt.Sprintf("%s-%d", attack.Topic, partition)
		for _, dir := range logDirs {
			entries, err := os.ReadDir(strings.TrimSpace(dir))
			if err != nil {
				return nil, perr.Wrapf(err, "read dir: %s", dir)
			}
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() == dirName {
					dirs = append(dirs, path.Join(strings.TrimSpace(dir), entry.Name()))
				}
			}

		}
	}
	return dirs, nil
}
