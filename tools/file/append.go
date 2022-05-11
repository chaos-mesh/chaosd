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

func NewFileAppendCommand() *cobra.Command {
	var fileName, data string
	var count int

	cmd := &cobra.Command{
		Use:   "append",
		Short: "append data to a file",

		Run: func(*cobra.Command, []string) {
			exit(appendFile(fileName, data, count))
		},
	}

	cmd.Flags().StringVarP(&fileName, "file-name", "f", "", "append data to the file")
	cmd.Flags().StringVarP(&data, "data", "d", "", "the appended data")
	cmd.Flags().IntVarP(&count, "count", "c", 1, "the append count of data")

	return cmd
}

func appendFile(fileName, data string, count int) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error("close file failed", zap.Error(err))
		}
	}()

	for i := 0; i < count; i++ {
		if _, err := f.Write([]byte(fmt.Sprintf("%s\n", data))); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
