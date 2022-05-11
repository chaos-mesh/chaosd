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
	"os"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewFileOrDirCreateCommand() *cobra.Command {
	var fileName, dirName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create file or directory",

		Run: func(*cobra.Command, []string) {
			exit(createFileOrDir(fileName, dirName))
		},
	}

	cmd.Flags().StringVarP(&fileName, "file-name", "f", "", "the name of created file")
	cmd.Flags().StringVarP(&dirName, "dir-name", "d", "", "the name of created directory")

	return cmd
}

func createFileOrDir(fileName, dirName string) error {
	var err error
	if len(fileName) > 0 {
		_, err = os.Create(fileName)
	} else if len(dirName) > 0 {
		err = os.Mkdir(dirName, os.ModePerm)
	}

	if err != nil {
		log.Error("create file/directory failed", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}
