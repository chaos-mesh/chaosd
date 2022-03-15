// Copyright 2021 Chaos Mesh Authors.
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
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

type completionContext struct {
	uids chan string
	err  chan error
}

func newCompletionCtx() *completionContext {
	return &completionContext{
		uids: make(chan string),
		err:  make(chan error),
	}
}

func listUid(ctx *completionContext, chaos *chaosd.Server) {
	exps, err := chaos.Search(&core.SearchCommand{All: true})
	if err != nil {
		ctx.err <- errors.Wrap(err, "list exp")
		return
	}

	for _, exp := range exps {
		ctx.uids <- exp.Uid
	}
	close(ctx.uids)
}
