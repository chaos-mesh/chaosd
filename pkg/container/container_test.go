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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
)

type dockerStub struct {
	inspectID, killID, signal string
	calls                     int
	err                       error
}

func (s *dockerStub) ContainerInspect(_ context.Context, id string) (dockertypes.ContainerJSON, error) {
	s.inspectID = id
	s.calls++
	return dockertypes.ContainerJSON{ContainerJSONBase: &dockertypes.ContainerJSONBase{State: &dockercontainer.State{Pid: 4321}}}, s.err
}

func (s *dockerStub) ContainerKill(_ context.Context, id, signal string) error {
	s.killID, s.signal = id, signal
	s.calls++
	return s.err
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

type dockerTransport func(*http.Request) (*http.Response, error)

func (f dockerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDockerSDKCompatibility(t *testing.T) {
	fixture := `{
		"Id":"abc123","Name":"/demo","State":{"Pid":4321,"Status":"running"},
		"Config":{"Image":"alpine","Env":["A=B"]},"HostConfig":{"KernelMemory":1048576},
		"Mounts":[{"Type":"bind","Source":"/source","Destination":"/target","RW":true}],
		"NetworkSettings":{"Bridge":"docker0","IPAddress":"172.17.0.2",
			"Ports":{"80/tcp":[{"HostIp":"127.0.0.1","HostPort":"8080"}]},
			"Networks":{"bridge":{"NetworkID":"net1","IPAddress":"172.17.0.2"}}}
	}`
	var expected dockertypes.ContainerJSON
	if err := json.Unmarshal([]byte(fixture), &expected); err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"", "1.44"} {
		t.Run("version="+version, func(t *testing.T) {
			pings, inspections, kills := 0, 0, 0
			transport := dockerTransport(func(req *http.Request) (*http.Response, error) {
				if req.URL.Host != "docker.test:2375" || req.Header.Get("X-Chaosd-Test") != "retained" {
					t.Fatalf("host=%q headers=%v", req.URL.Host, req.Header)
				}
				status, body := http.StatusOK, fixture
				headers := http.Header{"Content-Type": {"application/json"}}
				switch req.URL.Path {
				case "/_ping":
					pings++
					if req.Method != http.MethodHead {
						t.Fatalf("ping method=%s", req.Method)
					}
					headers.Set("API-Version", "1.44")
					body = ""
				case "/v1.44/containers/abc123/json":
					inspections++
					if req.Method != http.MethodGet || req.URL.RawQuery != "" {
						t.Fatalf("inspect request=%s %s", req.Method, req.URL)
					}
				case "/v1.44/containers/abc123/kill":
					kills++
					if req.Method != http.MethodPost || req.URL.Query().Get("signal") != "SIGTERM" {
						t.Fatalf("kill request=%s %s", req.Method, req.URL)
					}
					status, body = http.StatusNoContent, ""
				default:
					t.Fatalf("unexpected request=%s %s", req.Method, req.URL)
				}
				return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})
			client, err := newDockerClient("tcp://docker.test:2375", version, &http.Client{Transport: transport}, map[string]string{"X-Chaosd-Test": "retained"})
			if err != nil {
				t.Fatal(err)
			}
			actual, err := client.ContainerInspect(context.Background(), "abc123")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("inspect response=%+v, want %+v", actual, expected)
			}
			if actual.NetworkSettings.Bridge != "docker0" || actual.NetworkSettings.IPAddress != "172.17.0.2" || actual.HostConfig.KernelMemory != 1048576 {
				t.Fatal("inspect lost legacy response fields")
			}
			if err := client.ContainerKill(context.Background(), "abc123", "SIGTERM"); err != nil {
				t.Fatal(err)
			}
			wantPings := 0
			if version == "" {
				wantPings = 1
			}
			if pings != wantPings || inspections != 1 || kills != 1 {
				t.Fatalf("requests ping=%d inspect=%d kill=%d", pings, inspections, kills)
			}
		})
	}
}

func TestDockerSDKErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"not found", http.StatusNotFound, `{"message":"No such container"}`},
		{"malformed JSON", http.StatusOK, `{"Id":`},
		{"legacy field type", http.StatusOK, `{"Id":"abc123","NetworkSettings":{"Bridge":42}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport := dockerTransport(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: test.status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(test.body)), Request: req}, nil
			})
			client, err := newDockerClient("tcp://docker.test:2375", "1.44", &http.Client{Transport: transport}, nil)
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.ContainerInspect(context.Background(), "abc123")
			if err == nil {
				t.Fatal("inspect accepted failed response")
			}
			if test.name == "legacy field type" && result.ID != "abc123" {
				t.Fatal("inspect discarded partially decoded legacy response")
			}
			if test.status == http.StatusNotFound {
				if !strings.Contains(err.Error(), "No such container") {
					t.Fatalf("inspect error=%v", err)
				}
				if err := client.ContainerKill(context.Background(), "abc123", "SIGKILL"); err == nil || !strings.Contains(err.Error(), "No such container") {
					t.Fatalf("kill error=%v", err)
				}
			}
		})
	}
}
