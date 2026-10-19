// Copyright 2026 Chaos Mesh Authors.
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

import "testing"

func TestRedisExpirationOptions(t *testing.T) {
	for _, option := range []string{"", "NX", "XX", "GT", "LT", "invalid"} {
		t.Run(option, func(t *testing.T) {
			request := NewRedisCommand()
			request.Action = RedisCacheExpirationAction
			request.Addr = "127.0.0.1:6379"
			request.Expiration = "1m"
			request.Option = option
			err := request.Validate()
			if option == "invalid" {
				if err == nil {
					t.Fatal("invalid expiration option accepted")
				}
			} else if err != nil {
				t.Fatalf("valid expiration option rejected: %v", err)
			}
		})
	}
}

func TestRedisCacheLimitSelection(t *testing.T) {
	for _, size := range []string{"", "0", "1mb"} {
		t.Run(size, func(t *testing.T) {
			request := NewRedisCommand()
			request.Action = RedisCacheLimitAction
			request.Addr = "127.0.0.1:6379"
			request.Percent = "50%"
			request.CacheSize = size
			err := request.Validate()
			if size == "1mb" {
				if err == nil {
					t.Fatal("both size and percentage accepted")
				}
			} else if err != nil {
				t.Fatalf("percentage without an explicit size rejected: %v", err)
			}
		})
	}
}
