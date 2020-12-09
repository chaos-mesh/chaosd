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

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

var sFlag core.SearchCommand

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search UID",
		Short: "Search chaos attack, you can search attacks through the uid or the state of the attack",
		Run:   searchCommandFunc,
	}

	cmd.Flags().BoolVarP(&sFlag.All, "all", "A", false, "list all chaos attacks")
	cmd.Flags().StringVarP(&sFlag.Status, "status", "s", "", "attack status, "+
		"supported value: created, success, error, destroyed, revoked")
	cmd.Flags().StringVarP(&sFlag.Type, "type", "t", "", "attack type, "+
		"supported value: network, process")
	cmd.Flags().Uint32VarP(&sFlag.Offset, "offset", "o", 0, "starting to search attacks from offset")
	cmd.Flags().Uint32VarP(&sFlag.Limit, "limit", "l", 0, "limit the count of attacks")
	cmd.Flags().BoolVar(&sFlag.Asc, "asc", false, "order by CreateTime, "+
		"default value is false that means order by CreateTime desc")

	return cmd
}

func searchCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		ExitWithMsg(ExitBadArgs, "UID is required")
	}
	sFlag.UID = args[0]

	if err := sFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	exps, err := chaos.Search(&sFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"UID", "Type", "Action", "Status", "Create Time", "Configuration"})

	for _, exp := range exps {
		tw.AppendRow(table.Row{
			exp.Uid, exp.Kind, exp.Action, exp.Status, exp.CreatedAt, exp.RecoverCommand,
		})
	}

	tw.Style().Options.SeparateColumns = true

	fmt.Println(tw.Render())
}
