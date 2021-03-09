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

package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var hFlag core.HostCommand

func NewHostAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host <subcommand>",
		Short: "Host attack related commands",
	}

	cmd.AddCommand(NewHostShutdownCommand())

	return cmd
}

func NewHostShutdownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shutdown",
		Short: "shutdowns system, this action will trigger shutdown of the host machine",

		Run: processShutdownCommandFunc,
	}

	return cmd
}

func processShutdownCommandFunc(cmd *cobra.Command, args []string) {
	hFlag.Action = core.HostShutdownAction
	hostAttackF(cmd, &hFlag)
}

func hostAttackF(cmd *cobra.Command, f *core.HostCommand) {
	if err := hFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	uid, err := chaos.HostAttack(f)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	NormalExit(fmt.Sprintf("Attack host successfully, uid: %s", uid))
}
