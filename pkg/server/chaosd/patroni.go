// Copyright 2023 Chaos Mesh Authors.
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
)

type patroniAttack struct{}

var PatroniAttack AttackType = patroniAttack{}

func (patroniAttack) Attack(options core.AttackConfig, _ Environment) error {
	attack := options.(*core.PatroniCommand)

	candidate := attack.Candidate

	leader := attack.Leader

	var scheduled_at string

	var url string

	values := make(map[string]string)

	patroniInfo, err := utils.GetPatroniInfo(attack.Address)
	if err != nil {
		err = errors.Errorf("failed to get patroni info for %v: %v", options.String(), err)
		return errors.WithStack(err)
	}

	if len(patroniInfo.Replicas) == 0 {
		err = errors.Errorf("failed to get available replicas. Please, check your cluster")
		return errors.WithStack(err)
	}

	if candidate == "" {
		candidate = patroniInfo.Replicas[rand.Intn(len(patroniInfo.Replicas))]
	}

	if leader == "" {
		leader = patroniInfo.Master
	}

	switch options.String() {
	case "switchover":

		scheduled_at = attack.Scheduled_at

		values = map[string]string{"leader": leader, "scheduled_at": scheduled_at}

		log.Info(fmt.Sprintf("Switchover will be done from %v to another available replica in %v", patroniInfo.Master, scheduled_at))

	case "failover":

		values = map[string]string{"candidate": candidate}

		log.Info(fmt.Sprintf("Failover will be done from %v to %v", patroniInfo.Master, candidate))

	}

	patroniAddr := attack.Address

	cmd := options.String()

	data, err := json.Marshal(values)
	if err != nil {
		err = errors.Errorf("failed to marshal data: %v", values)
		return errors.WithStack(err)
	}

	url = fmt.Sprintf("http://%v:8008/%v", patroniAddr, cmd)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		err = errors.Errorf("failed to %v: %v", cmd, err)
		return errors.WithStack(err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(attack.User, attack.Password)

	client := &http.Client{}
	resp, error := client.Do(request)
	if error != nil {
		err = errors.Errorf("failed to %v: %v", cmd, err)
		return errors.WithStack(err)
	}

	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Errorf("failed to read %v responce: %v", cmd, err)
		return errors.WithStack(err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		err = errors.Errorf("failed to %v: status code %v, responce %v", cmd, resp.StatusCode, string(buf))
		return errors.WithStack(err)
	}

	log.S().Infof("Execute %v successfully: %v", cmd, string(buf))

	return nil
}

func (patroniAttack) Recover(exp core.Experiment, _ Environment) error {
	return nil
}
