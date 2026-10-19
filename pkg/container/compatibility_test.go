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

package container_test

import (
	"context"

	"github.com/docker/docker/api/types"

	"github.com/chaos-mesh/chaosd/pkg/container"
)

type existingClient struct{}

func (existingClient) ContainerInspect(context.Context, string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, nil
}
func (existingClient) ContainerKill(context.Context, string, string) error { return nil }

var _ container.DockerClientInterface = existingClient{}
