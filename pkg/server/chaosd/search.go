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

package chaosd

import (
	"context"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

func (s *Server) Search(conds *core.SearchCommand) ([]*core.Experiment, error) {
	if len(conds.UID) > 0 {
		exp, err := s.expStore.FindByUid(context.Background(), conds.UID)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return []*core.Experiment{exp}, nil
	}

	exps, err := s.expStore.ListByConditions(context.Background(), conds)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return exps, nil
}
