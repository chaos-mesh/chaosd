// Copyright 2023 Chaos Mesh Authors.
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

package utils

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/pingcap/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCommandPools_Cancel(t *testing.T) {
	now := time.Now()
	cmdPools := NewCommandPools(context.Background(), nil, 1)
	var gErr []error
	runner := NewCommandRunner("sleep", []string{"10s"}).
		WithOutputHandler(func(output []byte, err error, _ chan interface{}) {
			if err != nil {
				log.Error(string(output), zap.Error(err))
				gErr = append(gErr, err)
			}
			log.Info(string(output))
		}, nil)
	cmdPools.Start(runner)
	cmdPools.Close()
	assert.Less(t, time.Since(now).Seconds(), 10.0)
	assert.Equal(t, 1, len(gErr))
}

func TestCommandPools_Deadline(t *testing.T) {
	now := time.Now()
	deadline := time.Now().Add(time.Millisecond * 50)
	cmdPools := NewCommandPools(context.Background(), &deadline, 1)
	var gErr []error
	runner := NewCommandRunner("sleep", []string{"10s"}).
		WithOutputHandler(func(output []byte, err error, _ chan interface{}) {
			if err != nil {
				log.Error(string(output), zap.Error(err))
				gErr = append(gErr, err)
			}
			log.Info(string(output))
		}, nil)
	cmdPools.Start(runner)
	cmdPools.Wait()
	assert.Less(t, math.Abs(float64(time.Since(now).Milliseconds()-50)), 10.0)
	assert.Equal(t, 1, len(gErr))

}

func TestCommandPools_Normal(t *testing.T) {
	now := time.Now()
	cmdPools := NewCommandPools(context.Background(), nil, 1)
	var gErr []error
	runner := NewCommandRunner("sleep", []string{"1s"}).
		WithOutputHandler(func(output []byte, err error, _ chan interface{}) {
			if err != nil {
				log.Error(string(output), zap.Error(err))
				gErr = append(gErr, err)
			}
			log.Info(string(output))
		}, nil)
	cmdPools.Start(runner)
	cmdPools.Wait()
	assert.Less(t, time.Since(now).Seconds(), 2.0)
	assert.Equal(t, 0, len(gErr))
}
