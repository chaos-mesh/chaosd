// Copyright 2021 Chaos Mesh Authors.
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

package scheduler

import (
	"log"
	"testing"
	"time"
)

func TestScheduler_CronDurationExceeded(t *testing.T) {
	diffDuration, _ := time.ParseDuration("30s")
	startedAt := time.Now().Add(-diffDuration)
	log.Println(startedAt)
	for _, test := range []struct {
		name     string
		duration time.Duration
		want     bool
	}{
		{
			name:     "not exceeded",
			duration: diffDuration * 2,
			want:     false,
		},
		{
			name:     "exceeded",
			duration: diffDuration / 2,
			want:     true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := hasCronDurationExceeded(startedAt, test.duration)
			if got != test.want {
				t.Errorf("wrong result. got: %v, want: %v", got, test.want)
			}
		})
	}
}
