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
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/client"
	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/container"
	"github.com/chaos-mesh/chaos-daemon/pkg/server/chaosd"
	"github.com/chaos-mesh/chaos-daemon/pkg/store/dbstore"
	"github.com/chaos-mesh/chaos-daemon/pkg/store/experiment"
)

func mustClientFromCmd(cmd *cobra.Command) *client.Client {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	return client.NewClient(client.Config{
		Addr: url,
	})
}

func mustChaosdFromCmd(cmd *cobra.Command, conf *config.Config) *chaosd.Server {
	db, err := dbstore.DryDBStore()
	if err != nil {
		ExitWithError(ExitError, err)
	}

	cli, err := container.NewCRIClient(conf)
	if err != nil {
		ExitWithError(ExitError, err)
	}

	return chaosd.NewServer(conf, experiment.NewStore(db), cli, bpm.NewBackgroundProcessManager())
}
