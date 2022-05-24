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

package attack

import (
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func NewAttackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attack <subcommand>",
		Short: "Attack related commands",
	}

	var uid string
	cmd.PersistentFlags().StringVarP(&uid, "uid", "", "", "the experiment ID")

	cmd.AddCommand(
		NewProcessAttackCommand(&uid),
		NewNetworkAttackCommand(&uid),
		NewStressAttackCommand(&uid),
		NewDiskAttackCommand(&uid),
		NewHostAttackCommand(&uid),
		NewJVMAttackCommand(&uid),
		NewClockAttackCommand(&uid),
		NewRedisAttackCommand(&uid),
		NewFileAttackCommand(&uid),
		NewVMAttackCommand(&uid),
	)

	return cmd
}

func SetScheduleFlags(cmd *cobra.Command, conf *core.SchedulerConfig) {
	cmd.Flags().StringVar(&conf.Duration, "duration", "",
		`Work duration of attacks.A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m".Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
}
