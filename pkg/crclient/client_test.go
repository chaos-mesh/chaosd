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

package crclient

import (
	"context"
	"testing"
)

func TestNodeCRClientUsesHostPID(t *testing.T) {
	const localPID = 4321
	client := NewNodeCRClient(localPID)
	for _, id := range []string{"", "docker://abc123", "containerd://abc123", "pod-uid"} {
		t.Run(id, func(t *testing.T) {
			pid, err := client.GetPidFromContainerID(context.Background(), id)
			if err != nil || pid != localPID {
				t.Fatalf("container pid=%d err=%v, want %d", pid, err, localPID)
			}
			pid, err = client.GetSandboxPidFromPodUID(context.Background(), id)
			if err != nil || pid != localPID {
				t.Fatalf("sandbox pid=%d err=%v, want %d", pid, err, localPID)
			}
		})
	}
}
