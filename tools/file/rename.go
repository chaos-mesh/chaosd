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

func NewFileOrDirRenameCommand() *cobra.Command {
	var oldName, newName string

	cmd := &cobra.Command{
		Use:   "rename",
		Short: "rename file or directory",

		Run: func(*cobra.Command, []string) {
			exit(renameFileOrDir(oldName, newName))
		},
	}

	cmd.Flags().StringVarP(&oldName, "old-name", "o", "", "the old name of the file/directory")
	cmd.Flags().StringVarP(&newName, "new-name", "n", "", "the new name of the file/directory to be renamed")

	return cmd
}

func renameFileOrDir(oldName, newName string) error {
	err := os.Rename(oldName, newName)
	if err != nil {
		log.Error("rename file/dir faild", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}
