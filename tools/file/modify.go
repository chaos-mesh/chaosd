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

func NewFileModifyPrivilegeCommand() *cobra.Command {
	var fileName string
	var privilege uint32

	cmd := &cobra.Command{
		Use:   "modify",
		Short: "modify file's privilege",

		Run: func(*cobra.Command, []string) {
			modifyFilePrivilege(fileName, privilege)
		},
	}

	cmd.Flags().StringVarP(&fileName, "file-name", "f", "", "the name of the file")
	cmd.Flags().Uint32VarP(&privilege, "privilege", "p", 0, "the privilege of the file to be changed to, for example 777")

	return cmd
}

func modifyFilePrivilege(fileName string, privilege uint32) error {
	cmdStr := fmt.Sprintf("chmod %d %s", privilege, fileName)
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
