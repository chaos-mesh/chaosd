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
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
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
	cmd.Flags().StringVarP(&sFlag.Kind, "kind", "k", "", "attack kind, "+
		"supported value: network, process")
	cmd.Flags().Uint32VarP(&sFlag.Offset, "offset", "o", 0, "starting to search attacks from offset")
	cmd.Flags().Uint32VarP(&sFlag.Limit, "limit", "l", 0, "limit the count of attacks")
	cmd.Flags().BoolVar(&sFlag.Asc, "asc", false, "order by CreateTime, "+
		"default value is false that means order by CreateTime desc")

	return cmd
}

func searchCommandFunc(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		sFlag.UID = args[0]
	}

	if err := sFlag.Validate(); err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	chaos := mustChaosdFromCmd(cmd, &conf)

	exps, err := chaos.Search(&sFlag)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"UID", "Kind", "Action", "Status", "Create Time", "Configuration"})
	tw.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	tw.SetAlignment(3)
	tw.SetRowSeparator("-")
	tw.SetCenterSeparator(" ")
	tw.SetColumnSeparator(" ")

	for _, exp := range exps {
		tw.Append([]string{
			exp.Uid, exp.Kind, exp.Action, exp.Status, exp.CreatedAt.Format(time.RFC3339), exp.RecoverCommand,
		})
	}

	tw.Render()
}
