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

package main

import (
	"fmt"
	"os/exec"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewFileReplaceCommand() *cobra.Command {
	var file, originStr, destStr string
	var line int

	cmd := &cobra.Command{
		Use:   "replace",
		Short: "replace data in file",

		Run: func(*cobra.Command, []string) {
			exit(replaceFile(file, originStr, destStr, line))
		},
	}

	cmd.Flags().StringVarP(&file, "file-name", "f", "", "replace data in the file")
	cmd.Flags().StringVarP(&originStr, "origin-string", "o", "", "the origin string to be replaced")
	cmd.Flags().StringVarP(&destStr, "dest-string", "d", "", "the destination string to replace the origin string")
	cmd.Flags().IntVarP(&line, "line", "l", 0, "the line number to replace, default is 0, means replace all lines")

	return cmd
}

func replaceFile(file, originStr, destStr string, line int) error {
	var cmdStr string
	if line == 0 {
		cmdStr = fmt.Sprintf("sed -i 's/%s/%s/g' %s", originStr, destStr, file)
	} else {
		cmdStr = fmt.Sprintf("sed -i '%d s/%s/%s/g' %s", line, originStr, destStr, file)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return errors.WithStack(err)
	}
	if len(output) > 0 {
		log.Info(string(output))
	}

	return nil
}
