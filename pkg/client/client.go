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
	"net/http"
)

// Client is used to communicate with the chaosd
type Client struct {
	cfg    Config
	client *http.Client
}

// Config defines for chaosd client
type Config struct {
	Addr string
}

// NewClient creates a new chaosd client from a given address
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		client: http.DefaultClient,
	}
}
