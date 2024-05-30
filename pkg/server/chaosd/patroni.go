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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/utils"
)

type patroniAttack struct{}

var PatroniAttack AttackType = patroniAttack{}

func (patroniAttack) Attack(options core.AttackConfig, _ Environment) error {
	attack := options.(*core.PatroniCommand)

	var responce []byte

	var address string

	var err error

	values := make(map[string]string)

	if attack.RemoteMode {
		address = attack.Address
	} else if attack.LocalMode {
		address, err = utils.GetLocalHostname()
		if err != nil {
			return errors.WithStack(err)
		}
	}

	patroniInfo, err := utils.GetPatroniInfo(address)
	if err != nil {
		err = errors.Errorf("failed to get patroni info for %v: %v", options.String(), err)
		return errors.WithStack(err)
	}

	if len(patroniInfo.Replicas) == 0 && len(patroniInfo.SyncStandby) == 0 {
		err = errors.Errorf("failed to get available candidates. Please, check your cluster")
		return errors.WithStack(err)
	}

	sync_mode_check, err := isSynchronousClusterMode(address, attack.User, attack.Password)
	if err != nil {
		err = errors.Errorf("failed to check cluster synchronous mode for %v: %v", options.String(), err)
		return errors.WithStack(err)
	}

	if attack.Candidate == "" {
		if sync_mode_check {
			values["candidate"] = patroniInfo.SyncStandby[rand.Intn(len(patroniInfo.SyncStandby))]
		} else {
			values["candidate"] = patroniInfo.Replicas[rand.Intn(len(patroniInfo.Replicas))]
		}

	}

	if attack.Leader == "" {
		values["leader"] = patroniInfo.Master
	}

	values["scheduled_at"] = attack.Scheduled_at

	cmd := options.String()

	switch cmd {
	case "switchover":

		log.Info(fmt.Sprintf("Switchover will be done from %v to %v in %v", values["leader"], values["candidate"], values["scheduled_at"]))

	case "failover":

		log.Info(fmt.Sprintf("Failover will be done from %v to %v", values["leader"], values["candidate"]))

	}

	if attack.RemoteMode {
		responce, err = execPatroniAttackByRemoteMode(address, attack.User, attack.Password, cmd, values)
		if err != nil {
			return err
		}
	} else if attack.LocalMode {
		responce, err = execPatroniAttackByLocalMode(cmd, values)
		if err != nil {
			return err
		}
	}

	if attack.RemoteMode {
		log.S().Infof("Execute %v successfully: %v", cmd, string(responce))
	}

	if attack.LocalMode {
		log.S().Infof("Execute %v successfully", cmd)
		fmt.Println(string(responce))
	}

	return nil
}

func execPatroniAttackByRemoteMode(patroniAddr string, user string, password string, cmd string, values map[string]string) ([]byte, error) {

	data, err := json.Marshal(values)
	if err != nil {
		err = errors.Errorf("failed to marshal data: %v", values)
		return nil, errors.WithStack(err)
	}

	buf, err := utils.MakeHTTPRequest(http.MethodPost, patroniAddr, 8008, cmd, data, user, password)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return buf, nil
}

func execPatroniAttackByLocalMode(cmd string, values map[string]string) ([]byte, error) {
	var cmdTemplate string

	if cmd == "failover" {
		cmdTemplate = fmt.Sprintf("patronictl %v --master %v --candidate %v --force", cmd, values["leader"], values["candidate"])
	} else if cmd == "switchover" {
		cmdTemplate = fmt.Sprintf("patronictl %v --master %v --candidate %v --scheduled %v --force", cmd, values["leader"], values["candidate"], values["scheduled_at"])
	}

	execCmd := exec.Command("bash", "-c", cmdTemplate)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		log.S().Errorf(fmt.Sprintf("failed to %v: %v", cmdTemplate, string(output)))
		return nil, err
	}

	if strings.Contains(string(output), "failed") {
		err = errors.New(string(output))
		return nil, err
	}

	return output, nil
}

func isSynchronousClusterMode(patroniAddr string, user string, password string) (bool, error) {

	buf, err := utils.MakeHTTPRequest(http.MethodGet, patroniAddr, 8008, "config", []byte{}, user, password)
	if err != nil {
		return false, err
	}

	patroni_responce := make(map[string]interface{})

	err = json.Unmarshal(buf, &patroni_responce)
	if err != nil {
		return false, fmt.Errorf("bad request %v %v", err.Error(), http.StatusBadRequest)
	}

	mode_check, ok := patroni_responce["synchronous_mode"].(bool)
	if !ok {
		return false, fmt.Errorf("failed to cast synchronous_mode field from patroni responce")
	}

	if mode_check {
		return true, nil
	}

	return false, nil

}

func (patroniAttack) Recover(exp core.Experiment, _ Environment) error {
	return nil
}
