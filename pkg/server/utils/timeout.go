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
	"os/exec"
	"time"

	"github.com/pingcap/log"
	"go.uber.org/zap"
)

func ExecWithDeadline(t <-chan time.Time, cmd *exec.Cmd) error {
	done := make(chan error, 1)
	var output []byte
	var err error
	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-t:
		if err := cmd.Process.Kill(); err != nil {
			log.Error("failed to kill process: ", zap.Error(err))
			return err
		}
	case err := <-done:
		if err != nil {
			log.Error(err.Error()+string(output), zap.Error(err))
			return err
		}
		log.Info(string(output))
	}
	return nil
}
