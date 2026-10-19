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

package experiment

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/store/dbstore"
)

func TestExperimentOrderingAndScheduledSearch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "experiments.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	store := NewStore(&dbstore.DB{DB: db})
	runs := NewRunStore(&dbstore.DB{DB: db})
	ctx := context.Background()
	var newest *core.Experiment
	for _, sec := range []int64{1, 3, 2} {
		exp := &core.Experiment{Uid: "same-uid", CreatedAt: time.Unix(sec, 0), Status: core.Scheduled, LaunchMode: core.ServerMode}
		if err := store.Set(ctx, exp); err != nil {
			t.Fatal(err)
		}
		if sec == 3 {
			newest = exp
		}
	}
	other := &core.Experiment{Uid: "other", CreatedAt: time.Unix(4, 0), Status: core.Success, LaunchMode: core.CommandMode}
	if err := store.Set(ctx, other); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		get  func() ([]*core.Experiment, error)
		want []int64
	}{
		{"all", func() ([]*core.Experiment, error) { return store.List(ctx) }, []int64{4, 3, 2, 1}},
		{"launch-mode", func() ([]*core.Experiment, error) { return store.ListByLaunchMode(ctx, core.ServerMode) }, []int64{3, 2, 1}},
		{"status", func() ([]*core.Experiment, error) { return store.ListByStatus(ctx, core.Scheduled) }, []int64{3, 2, 1}},
		{"scheduled-search", func() ([]*core.Experiment, error) {
			return store.ListByConditions(ctx, &core.SearchCommand{Status: core.Scheduled})
		}, []int64{3, 2, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.get()
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d rows, want %d", len(got), len(tc.want))
			}
			for i, want := range tc.want {
				if got[i].CreatedAt.Unix() != want {
					t.Errorf("row %d timestamp = %d, want %d", i, got[i].CreatedAt.Unix(), want)
				}
			}
		})
	}
	t.Run("duplicate-uid-newest", func(t *testing.T) {
		got, err := store.FindByUid(ctx, "same-uid")
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != newest.ID {
			t.Fatalf("got ID %d, want newest ID %d", got.ID, newest.ID)
		}
	})
	for _, sec := range []int64{1, 3, 2} {
		if err := runs.NewRun(ctx, &core.ExperimentRun{UID: time.Unix(sec, 0).String(), ExperimentID: newest.ID, StartAt: time.Unix(sec, 0)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := runs.NewRun(ctx, &core.ExperimentRun{UID: "unrelated", ExperimentID: other.ID, StartAt: time.Unix(4, 0)}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		get  func() ([]*core.ExperimentRun, error)
	}{
		{"runs-by-id", func() ([]*core.ExperimentRun, error) { return runs.ListByExperimentID(ctx, newest.ID) }},
		{"runs-by-uid", func() ([]*core.ExperimentRun, error) { return runs.ListByExperimentUID(ctx, "same-uid") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.get()
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 3 {
				t.Fatalf("got %d runs", len(got))
			}
			for i, want := range []int64{3, 2, 1} {
				if got[i].StartAt.Unix() != want {
					t.Errorf("run %d timestamp = %d, want %d", i, got[i].StartAt.Unix(), want)
				}
				if got[i].Experiment.ID != newest.ID {
					t.Errorf("run %d missing associated experiment", i)
				}
			}
		})
	}
}

func TestStoreSchemaMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "schema.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&core.Experiment{}, &core.ExperimentRun{}, &core.IPSetRule{}, &core.IptablesRule{}, &core.TCRule{}); err != nil {
		t.Fatal(err)
	}
}
