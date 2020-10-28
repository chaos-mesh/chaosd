// Copyright 2020 Chaos Mesh Authors.
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

package ctl

import (
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-daemon/cmd/chaos/ctl/command"
)

type CommandFlags struct {
	URL string
}

// CommandFlags are flags that used in all Commands
var rootCmd = &cobra.Command{
	Use:   "chaos",
	Short: "A command line client to run chaos experiment",
}

var (
	cmdFlags = CommandFlags{}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cmdFlags.URL, "url", "u", "http://127.0.0.0:31767", "chaosd address")
	rootCmd.AddCommand(
		command.NewAttackCommand(),
	)
}

// Execute execs Command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		command.ExitWithError(command.ExitError, err)
	}
}
