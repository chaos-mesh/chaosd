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

package core

import (
	"github.com/pingcap/errors"
)

type SearchCommand struct {
	Asc    bool
	All    bool
	Status string
	Kind   string
	Limit  uint32
	Offset uint32
	UID    string
}

func (s *SearchCommand) Validate() error {
	if len(s.UID) > 0 {
		return nil
	}

	if len(s.Kind) > 0 {
		switch s.Kind {
		case NetworkAttack, ProcessAttack:
			break
		default:
			return errors.Errorf("type %s not supported", s.Kind)
		}
	}

	if len(s.Status) > 0 {
		switch s.Status {
		case Created, Success, Error, Destroyed, Revoked:
			break
		default:
			return errors.Errorf("status %s not supported", s.Status)
		}
	}

	if len(s.Status) == 0 && len(s.Kind) == 0 && !s.All {
		return errors.New("UID is required")
	}

	return nil
}
