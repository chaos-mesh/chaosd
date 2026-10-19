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

package container

import (
	"context"
	"errors"
	"testing"

	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

type dockerStub struct {
	inspectID, killID, signal string
	calls                     int
	err                       error
}

func (s *dockerStub) ContainerInspect(_ context.Context, id string, _ dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error) {
	s.inspectID = id
	s.calls++
	return dockerclient.ContainerInspectResult{Container: dockercontainer.InspectResponse{State: &dockercontainer.State{Pid: 4321}}}, s.err
}

func (s *dockerStub) ContainerKill(_ context.Context, id string, options dockerclient.ContainerKillOptions) (dockerclient.ContainerKillResult, error) {
	s.killID, s.signal = id, options.Signal
	s.calls++
	return dockerclient.ContainerKillResult{}, s.err
}

func TestDockerClientRequests(t *testing.T) {
	stub := &dockerStub{}
	client := DockerClient{client: stub}
	pid, err := client.GetPidFromContainerID(context.Background(), "docker://abc123")
	if err != nil || pid != 4321 || stub.inspectID != "abc123" {
		t.Fatalf("inspect returned pid=%d err=%v; requested id=%q", pid, err, stub.inspectID)
	}
	if err := client.ContainerKillByContainerID(context.Background(), "docker://abc123"); err != nil {
		t.Fatal(err)
	}
	if stub.killID != "abc123" || stub.signal != "SIGKILL" || stub.calls != 2 {
		t.Fatalf("kill requested id=%q signal=%q; total calls=%d", stub.killID, stub.signal, stub.calls)
	}
}

func TestDockerClientErrors(t *testing.T) {
	failure := errors.New("runtime unavailable")
	stub := &dockerStub{err: failure}
	client := DockerClient{client: stub}
	if _, err := client.GetPidFromContainerID(context.Background(), "docker://abc123"); !errors.Is(err, failure) {
		t.Fatalf("inspect error=%v, want runtime error", err)
	}
	if err := client.ContainerKillByContainerID(context.Background(), "docker://abc123"); !errors.Is(err, failure) {
		t.Fatalf("kill error=%v, want runtime error", err)
	}
	for _, id := range []string{"", "abc", "containerd://abc123", "docker:/abc123"} {
		t.Run(id, func(t *testing.T) {
			stub := &dockerStub{}
			client := DockerClient{client: stub}
			if _, err := client.GetPidFromContainerID(context.Background(), id); err == nil {
				t.Fatal("inspect accepted malformed prefix")
			}
			if err := client.ContainerKillByContainerID(context.Background(), id); err == nil {
				t.Fatal("kill accepted malformed prefix")
			}
			if stub.calls != 0 {
				t.Fatalf("malformed prefix made %d runtime calls", stub.calls)
			}
		})
	}
}
