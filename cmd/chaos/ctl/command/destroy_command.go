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

import "github.com/spf13/cobra"

func NewDestroyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy UID",
		Short: "Destroy a chaos experiment",
		Args:  cobra.MinimumNArgs(1),
		Run:   destroyCommandF,
	}

	return cmd
}

func destroyCommandF(cmd *cobra.Command, args []string) {

}
