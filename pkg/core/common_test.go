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

package core

import "testing"

func TestScheduleConfig_ParseDuration(t *testing.T) {
	t.Run("EmptyDurationConfig", func(t *testing.T) {
		cfg := SchedulerConfig{}
		dur, err := cfg.ScheduleDuration()
		if dur != nil || err != nil {
			t.Errorf("invalid result. duration: %v, error: %v", dur, err)
		}
	})
	t.Run("CorrectDurationConfig", func(t *testing.T) {
		cfg := SchedulerConfig{
			Duration: "15m",
		}
		dur, err := cfg.ScheduleDuration()
		if dur == nil || err != nil {
			t.Errorf("invalid result. duration: %v, error: %v", dur, err)
		}
	})
	t.Run("WrongDurationConfig", func(t *testing.T) {
		cfg := SchedulerConfig{
			Duration: "m15",
		}
		dur, err := cfg.ScheduleDuration()
		if err == nil {
			t.Errorf("invalid result. duration: %v, error: %v", dur, err)
		}
	})
}

func TestAttackDurationValidation(t *testing.T) {
	for _, tc := range []struct {
		duration, schedule string
		invalid            bool
	}{
		{"", "", false},
		{"100ms", "", false},
		{"invalid", "", false},
		{"", "* * * * *", true},
		{"100ms", "* * * * *", false},
		{"invalid", "* * * * *", true},
	} {
		config := CommonAttackConfig{Duration: tc.duration, Schedule: tc.schedule}
		if err := config.Validate(); (err != nil) != tc.invalid {
			t.Errorf("duration=%q schedule=%q: %v", tc.duration, tc.schedule, err)
		}
	}
}
