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
	"io"
	"os"

	"github.com/spf13/cobra"
)

func NewFileCopyCommand() *cobra.Command {
	var fileName, copyFileName string

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "copy file",

		Run: func(*cobra.Command, []string) {
			exit(copyFile(fileName, copyFileName))

		},
	}

	cmd.Flags().StringVarP(&fileName, "file-name", "f", "", "the name of the file to be copied")
	cmd.Flags().StringVarP(&copyFileName, "copy-file-name", "c", "", "the name of the file to be copied to")

	return cmd
}

func copyFile(fileName, copyFileName string) error {
	sourceFileStat, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", fileName)
	}

	source, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(copyFileName)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
