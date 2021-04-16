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

package search

import (
	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/utils"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

func NewSearchCommand() *cobra.Command {
	options := &core.SearchCommand{}
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *core.SearchCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "search UID",
		Short: "Search chaos attack, you can search attacks through the uid or the state of the attack",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				options.UID = args[0]
			}
			fx.New(dep, fx.Invoke(searchCommandFunc)).Run()
		},
	}

	cmd.Flags().BoolVarP(&options.All, "all", "A", false, "list all chaos attacks")
	cmd.Flags().StringVarP(&options.Status, "status", "s", "", "attack status, "+
		"supported value: created, success, error, destroyed, revoked")
	cmd.Flags().StringVarP(&options.Kind, "kind", "k", "", "attack kind, "+
		"supported value: network, process")
	cmd.Flags().Uint32VarP(&options.Offset, "offset", "o", 0, "starting to search attacks from offset")
	cmd.Flags().Uint32VarP(&options.Limit, "limit", "l", 0, "limit the count of attacks")
	cmd.Flags().BoolVar(&options.Asc, "asc", false, "order by CreateTime, "+
		"default value is false that means order by CreateTime desc")

	return cmd
}

func searchCommandFunc(chaos *chaosd.Server, options *core.SearchCommand) {
	if err := options.Validate(); err != nil {
		utils.ExitWithError(utils.ExitBadArgs, err)
	}

	exps, err := chaos.Search(options)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
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

	utils.NormalExit("")
}
