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

package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

const (
	processAttack = "api/attack/process"
)

func (c *Client) CreateProcessAttack(attack *core.ProcessCommand) (*utils.Response, *utils.APIError, error) {
	a, err := json.Marshal(attack)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	url := fmt.Sprintf("%s/%s", c.cfg.Addr, processAttack)
	data, apiErr, err := doRequest(c.client, url, http.MethodPost, withJsonBody(a))
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if apiErr != nil {
		aerr := &utils.APIError{}
		if err := json.Unmarshal(apiErr, aerr); err != nil {
			return nil, nil, errors.WithStack(err)
		}
		return nil, aerr, nil
	}

	resp := &utils.Response{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return resp, nil, nil
}
