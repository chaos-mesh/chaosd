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

package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/pingcap/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type PatroniInfo struct {
	Master      string
	Replicas    []string
	SyncStandby []string
	Status      []string
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
		switch member.Get("role").Str {
		case "leader":
			patroniInfo.Master = member.Get("name").Str
			patroniInfo.Status = append(patroniInfo.Status, member.Get("state").Str)
		case "replica":
			patroniInfo.Replicas = append(patroniInfo.Replicas, member.Get("name").Str)
			patroniInfo.Status = append(patroniInfo.Status, member.Get("state").Str)
		case "sync_standby":
			patroniInfo.SyncStandby = append(patroniInfo.SyncStandby, member.Get("name").Str)
			patroniInfo.Status = append(patroniInfo.Status, member.Get("state").Str)

		}
	}

	log.Info(fmt.Sprintf("patroni info: master %v, replicas %v, sync_standy %s, statuses %v\n", patroniInfo.Master, patroniInfo.Replicas,
		patroniInfo.SyncStandby, patroniInfo.Status))

	return patroniInfo, nil

}

func MakeHTTPRequest(method string, address string, port int64, path string, body []byte, user string, password string) ([]byte, error) {

	url := fmt.Sprintf("http://%v:%v/%v", address, port, path)

	var request *http.Request

	var resp *http.Response

	var err error

	switch method {
	case http.MethodPost:
		request, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
		if err != nil {
			err = errors.Errorf("failed to post request %v: %v", url, err)
			return nil, errors.WithStack(err)
		}

	case http.MethodGet:
		request, err = http.NewRequest("GET", url, nil)
		if err != nil {
			err = errors.Errorf("failed to get request %v: %v", url, err)
			return nil, errors.WithStack(err)
		}
	}

	if user != "" && password != "" {
		request.Header.Set("Content-Type", "application/json")
		request.SetBasicAuth(user, password)
	}

	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		err = errors.Errorf("failed to exec %v request %v: %v", method, url, err)
		return nil, errors.WithStack(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		//to simplify diagnostics
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			err = errors.Errorf("failed to read from %s responce: status code %v, responce %v, error %v", path, resp.StatusCode, resp.Body, err)
			return nil, err
		}
		err = errors.Errorf("failed to exec %v request: status code %v, responce %v", path, resp.StatusCode, buf)
		return nil, errors.WithStack(err)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Errorf("failed to read %v from %s responce: %v", resp.Body, path, err)
		return nil, errors.WithStack(err)
	}

	return buf, nil
}
