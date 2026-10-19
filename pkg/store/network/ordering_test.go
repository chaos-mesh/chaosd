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

package network

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

func TestRuleListsNewestFirst(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "rules.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	storage := &dbstore.DB{DB: db}
	ipsets := NewIPSetRuleStore(storage)
	iptables := NewIptablesRuleStore(storage)
	tcs := NewTCRuleStore(storage)
	for _, sec := range []int64{1, 3, 2} {
		model := gorm.Model{CreatedAt: time.Unix(sec, 0)}
		for _, row := range []interface{}{
			&core.IPSetRule{Model: model, Experiment: "target"},
			&core.IptablesRule{Model: model, Experiment: "target"},
			&core.TCRule{Model: model, Experiment: "target"},
		} {
			if err := db.Create(row).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		list func() ([]gorm.Model, error)
	}{
		{"ipsets", func() ([]gorm.Model, error) {
			rows, err := ipsets.List(ctx)
			var result []gorm.Model
			for _, row := range rows {
				result = append(result, row.Model)
			}
			return result, err
		}},
		{"ipsets-by-experiment", func() ([]gorm.Model, error) {
			rows, err := ipsets.FindByExperiment(ctx, "target")
			var result []gorm.Model
			for _, row := range rows {
				result = append(result, row.Model)
			}
			return result, err
		}},
		{"iptables", func() ([]gorm.Model, error) {
			rows, err := iptables.List(ctx)
			var result []gorm.Model
			for _, row := range rows {
				result = append(result, row.Model)
			}
			return result, err
		}},
		{"traffic-control", func() ([]gorm.Model, error) {
			rows, err := tcs.List(ctx)
			var result []gorm.Model
			for _, row := range rows {
				result = append(result, row.Model)
			}
			return result, err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := tc.list()
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 3 {
				t.Fatalf("got %d rows", len(rows))
			}
			for i, sec := range []int64{3, 2, 1} {
				if rows[i].CreatedAt.Unix() != sec {
					t.Errorf("row %d is out of order", i)
				}
			}
		})
	}
}

func TestMigrateLegacyNetworkIndexes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&core.IptablesRule{}); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE INDEX name ON iptables_rules(name)",
		"CREATE INDEX experiment ON iptables_rules(experiment)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	old := &core.IptablesRule{Name: "preserved-chain", IPSets: "preserved-set", Direction: "INPUT", Experiment: "legacy", Protocol: "tcp"}
	if err := db.Create(old).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&core.IPSetRule{}, &core.IptablesRule{}, &core.TCRule{}); err != nil {
		t.Fatal(err)
	}
	var got core.IptablesRule
	if err := db.First(&got, old.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.ID != old.ID || got.Name != old.Name || got.IPSets != old.IPSets || got.Direction != old.Direction || got.Experiment != old.Experiment || got.Protocol != old.Protocol || !got.CreatedAt.Equal(old.CreatedAt) || !got.UpdatedAt.Equal(old.UpdatedAt) || got.DeletedAt != old.DeletedAt {
		t.Fatalf("legacy rule changed: got %+v, want %+v", got, *old)
	}
	for _, statement := range []string{
		"SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'name' AND tbl_name = 'iptables_rules'",
		"SELECT name FROM sqlite_master WHERE type = 'index' AND name = 'experiment' AND tbl_name = 'iptables_rules'",
	} {
		var name string
		if err := db.Raw(statement).Scan(&name).Error; err != nil {
			t.Fatal(err)
		}
		if name == "" {
			t.Fatal("legacy index was removed")
		}
	}
}
