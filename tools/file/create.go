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
	"os"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewFileOrDirCreateCommand() *cobra.Command {
	var fileName, dirName, destDir string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create file or directory",

		Run: func(*cobra.Command, []string) {
			createFileOrDir(fileName, dirName, destDir)
		},
	}

	cmd.Flags().StringVarP(&fileName, "file-name", "f", "", "the name of created file")
	cmd.Flags().StringVarP(&dirName, "dir-name", "d", "", "the name of created directory")
	cmd.Flags().StringVarP(&destDir, "dest-dir", "", "", "create a file or directory based on the specified directory")

	return cmd
}

func createFileOrDir(fileName, dirName, destDir string) error {
	var err error
	if len(fileName) > 0 {
		_, err = os.Create(fmt.Sprintf("%s/%s", destDir, fileName))
	} else if len(dirName) > 0 {
		err = os.Mkdir(fmt.Sprintf("%s/%s", destDir, dirName), os.ModePerm)
	}

	if err != nil {
		log.Error("create file/directory failed", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}
