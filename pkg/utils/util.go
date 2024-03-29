// Copyright 2021 Chaos Mesh Authors.
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
	"math/rand"
	"os/exec"
	"time"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandomStringWithCharset(length int) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func ExecuteCmd(cmdStr string) (string, error) {
	log.Info("execute cmd", zap.String("cmd", cmdStr))
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return "", err
	}
	if len(output) > 0 {
		log.Info("command output: "+string(output), zap.String("command", cmdStr))
	}

	return string(output), nil
}
