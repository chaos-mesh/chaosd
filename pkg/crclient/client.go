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

package crclient

import "context"

func NewNodeCRClient(pid int) *NodeCRClient {
	return &NodeCRClient{
		Pid: uint32(pid),
	}
}

type NodeCRClient struct {
	Pid uint32
}

func (n *NodeCRClient) GetPidFromContainerID(_ context.Context, _ string) (uint32, error) {
	return n.Pid, nil
}

func (n *NodeCRClient) ContainerKillByContainerID(_ context.Context, _ string) error {
	return nil
}

func (n *NodeCRClient) FormatContainerID(_ context.Context, _ string) (string, error) {
	return "", nil
}
