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
	"bytes"
	"io"
	"net/http"

	"github.com/pingcap/errors"
)

// BodyOption sets the type and content of the body
type BodyOption func(*bodyOption)

type bodyOption struct {
	contentType string
	body        io.Reader
}

func withJsonBody(body []byte) BodyOption {
	return func(bo *bodyOption) {
		bo.contentType = "application/json"
		bo.body = bytes.NewBuffer(body)
	}
}

func doRequest(
	cli *http.Client,
	url, method string,
	opts ...BodyOption,
) ([]byte, []byte, error) {
	b := &bodyOption{}
	for _, o := range opts {
		o(b)
	}
	req, err := http.NewRequest(method, url, b.body)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if b.contentType != "" {
		req.Header.Set("Content-Type", b.contentType)
	}

	resp, apiErr, err := dial(cli, req)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return resp, apiErr, nil
}

func dial(cli *http.Client, req *http.Request) ([]byte, []byte, error) {
	resp, err := cli.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiErr []byte
		apiErr, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, apiErr, nil
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return content, nil, nil
}
