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
	"context"
	"fmt"

	"github.com/pingcap/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all the attack experiments",
		Run:   listExpFunc,
	}

	return cmd
}

func listExpFunc(cmd *cobra.Command, args []string) {
	chaos := mustChaosdFromCmd(cmd, &conf)
	exps, err := chaos.Exp().List(context.Background())
	if err != nil {
		log.Error("failed to list experiments", zap.Error(err))
		return
	}

	for _, exp := range exps {
		fmt.Println(exp)
	}
}
