// Copyright 2023 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.
package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pingcap/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type PatroniInfo struct {
	Master   string
	Replicas []string
	Status   []string
}

func GetPatroniInfo(address string) (PatroniInfo, error) {
	res, err := http.Get(fmt.Sprintf("http://%v:8008/cluster", address))
	if err != nil {
		err = errors.Errorf("failed to get patroni status: %v", err)
		return PatroniInfo{}, errors.WithStack(err)
	}

	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		err = errors.Errorf("failed to read responce: %v", err)
		return PatroniInfo{}, errors.WithStack(err)
	}

	data := string(buf)

	patroniInfo := PatroniInfo{}

	members := gjson.Get(data, "members")

	for _, member := range members.Array() {
		if member.Get("role").Str == "leader" {
			patroniInfo.Master = member.Get("name").Str
			patroniInfo.Status = append(patroniInfo.Status, member.Get("state").Str)
		} else if member.Get("role").Str == "replica" || member.Get("role").Str == "sync_standby" {
			patroniInfo.Replicas = append(patroniInfo.Replicas, member.Get("name").Str)
			patroniInfo.Status = append(patroniInfo.Status, member.Get("state").Str)
		}
	}

	log.Info(fmt.Sprintf("patroni info: master %v, replicas %v, statuses %v\n", patroniInfo.Master, patroniInfo.Replicas, patroniInfo.Status))

	return patroniInfo, nil

}
