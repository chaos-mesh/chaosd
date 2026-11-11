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

package recover

import (
	"context"
	"fmt"
	"os"

	"github.com/joomcode/errorx"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/cmd/server"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type recoverCommand struct {
	uid  string
	all  bool
	kind string
}

func (r *recoverCommand) validate() error {
	recoverMany := r.all || len(r.kind) > 0
	if len(r.uid) > 0 && recoverMany {
		return errors.New("UID cannot be used with --all or --kind")
	}

	if len(r.uid) == 0 && !recoverMany {
		return errors.New("UID, --all or --kind is required")
	}

	if len(r.kind) > 0 && core.GetAttackByKind(r.kind) == nil {
		return errors.Errorf("kind %s not supported", r.kind)
	}

	return nil
}

func NewRecoverCommand() *cobra.Command {
	options := &recoverCommand{}
	dep := fx.Options(
		server.Module,
		fx.Provide(func() *recoverCommand {
			return options
		}),
	)

	cmd := &cobra.Command{
		Use:   "recover [UID]",
		Short: "Run recovery for an experiment by UID, or for all injected experiments",
		Long: `Run recovery for an experiment by UID, or for all injected experiments.

With --all or --kind, experiments in the success status are recovered from newest to oldest.
Experiments whose faults cannot be recovered are skipped, and a failed recovery does not stop the others.
Scheduled experiments are not included; cancel them by UID through the service that owns their schedule.`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeUID,
		Run: func(_ *cobra.Command, args []string) {
			if len(args) > 0 {
				options.uid = args[0]
			}
			if err := options.validate(); err != nil {
				utils.ExitWithError(utils.ExitBadArgs, err)
			}

			if len(options.uid) > 0 {
				utils.FxNewAppWithoutLog(dep, fx.Invoke(recoverCommandF)).Run()
			} else {
				utils.FxNewAppWithoutLog(dep, fx.Invoke(recoverAllCommandF)).Run()
			}
		},
	}

	cmd.Flags().BoolVarP(&options.all, "all", "A", false, "Recover all injected experiments")
	cmd.Flags().StringVarP(&options.kind, "kind", "k", "", "Recover all injected experiments of a kind: network, process, stress, disk, disk-server, host, jvm, clock, kafka, redis, file, http, vm, or userDefined")
	cmd.MarkFlagsMutuallyExclusive("all", "kind")

	return cmd
}

func recoverCommandF(chaos *chaosd.Server, options *recoverCommand) {
	err := chaos.RecoverAttack(options.uid)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	utils.NormalExit(fmt.Sprintf("Recover %s successfully", options.uid))
}

func recoverAllCommandF(chaos *chaosd.Server, options *recoverCommand) {
	results, err := chaos.RecoverAttacks(options.kind)
	if err != nil {
		utils.ExitWithError(utils.ExitError, err)
	}

	var recovered, skipped, failed int
	for _, result := range results {
		exp := result.Experiment
		switch {
		case result.Err == nil:
			recovered++
			fmt.Printf("Recover %s successfully\n", exp.Uid)
		case errorx.IsOfType(result.Err, core.ErrNonRecoverableAttack):
			skipped++
			fmt.Printf("Skip %s: %s %s fault cannot be recovered\n", exp.Uid, exp.Kind, exp.Action)
		default:
			failed++
			fmt.Fprintln(os.Stderr, "Error:", result.Err)
		}
	}

	summary := fmt.Sprintf("Recovered %d, skipped %d, failed %d experiments", recovered, skipped, failed)
	if failed > 0 {
		utils.ExitWithMsg(utils.ExitError, summary)
	}
	utils.NormalExit(summary)
}

func completeUID(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	completionCtx := newCompletionCtx()
	completionDep := fx.Options(
		server.Module,
		fx.Provide(func() *completionContext {
			return completionCtx
		}),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := utils.FxNewAppWithoutLog(completionDep, fx.Invoke(listUID)).Start(ctx); err != nil {
			log.Error(errors.Wrap(err, "start application").Error())
		}
	}()
	var uids []string
	for {
		select {
		case uid := <-completionCtx.uids:
			if len(uid) == 0 {
				return uids, cobra.ShellCompDirectiveNoFileComp
			}
			uids = append(uids, uid)
		case err := <-completionCtx.err:
			log.Error(err.Error())
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}
