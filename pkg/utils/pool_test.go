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

var GlobalErrors []error

func TestCommandPools_Cancel(t *testing.T) {
	now := time.Now()
	cmdPools := NewCommandPools(context.Background(), nil, 1)
	cmdPools.Start("sleep", []string{"10s"}, func(output []byte, err error) {
		if err != nil {
			log.Error(string(output), zap.Error(err))
			GlobalErrors = append(GlobalErrors, err)
		}
		log.Info(string(output))
	})
	cmdPools.Close()
	assert.Less(t, time.Since(now).Seconds(), 10.0)
	assert.Equal(t, 1, len(GlobalErrors))
	GlobalErrors = []error{}
}

func TestCommandPools_Deadline(t *testing.T) {
	now := time.Now()
	deadline := time.Now().Add(time.Millisecond * 50)
	cmdPools := NewCommandPools(context.Background(), &deadline, 1)
	cmdPools.Start("sleep", []string{"10s"}, func(output []byte, err error) {
		if err != nil {
			log.Error(string(output), zap.Error(err))
			GlobalErrors = append(GlobalErrors, err)
		}
		log.Info(string(output))
	})
	cmdPools.Wait()
	assert.Less(t, math.Abs(float64(time.Since(now).Milliseconds()-50)), 10.0)
	assert.Equal(t, 1, len(GlobalErrors))
	GlobalErrors = []error{}
}

func TestCommandPools_Normal(t *testing.T) {
	now := time.Now()
	cmdPools := NewCommandPools(context.Background(), nil, 1)
	cmdPools.Start("sleep", []string{"1s"}, func(output []byte, err error) {
		if err != nil {
			log.Error(string(output), zap.Error(err))
			GlobalErrors = append(GlobalErrors, err)
		}
		log.Info(string(output))
	})
	cmdPools.Wait()
	assert.Less(t, time.Since(now).Seconds(), 2.0)
	assert.Equal(t, 0, len(GlobalErrors))
	GlobalErrors = []error{}
}
