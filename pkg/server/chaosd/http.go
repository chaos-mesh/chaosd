// Copyright 2022 Chaos Mesh Authors.
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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/pkg/errors"
	"net/http"
	"os/exec"
)

type attackHTTP struct{}

var AttackHTTP AttackType = attackHTTP{}

func (attackHTTP) Attack(options core.AttackConfig, env Environment) error {
	var attackConf *core.HTTPAttackConfig
	var ok bool
	if attackConf, ok = options.(*core.HTTPAttackConfig); !ok {
		return fmt.Errorf("AttackConfig -> *HTTPAttackConfig meet error")
	}

	cmd := exec.Command("tproxy", "-i", "-vv")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "create stdin pipe")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "create stdout pipe")
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "start command `%s`", cmd.String())
	}

	config, err := json.Marshal(&attackConf)
	if err != nil {
		return errors.Wrap(err, "applying HTTP attack")
	}

	req, err := http.NewRequest(http.MethodPut, "/", bytes.NewReader(config))
	if err != nil {
		return errors.Wrap(err, "create http request")
	}

	err = req.Write(stdin)
	if err != nil {
		return errors.Wrap(err, "cannot request tproxy")
	}

	resp, err := http.ReadResponse(bufio.NewReader(stdout), req)

	fmt.Println(resp)

	err = cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "waiting cmd")
	}
	return nil
}

func (attackHTTP) Recover(experiment core.Experiment, env Environment) error {
	//TODO implement me
	panic("implement me")
}
