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

package chaosd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/scheduler"
	"github.com/chaos-mesh/chaosd/pkg/store/dbstore"
	"github.com/chaos-mesh/chaosd/pkg/store/experiment"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func TestDiskPayloadDuration(t *testing.T) {
	for _, server := range []bool{false, true} {
		name := "command"
		if server {
			name = "server"
		}
		for _, duration := range []string{"100ms", ""} {
			t.Run(name+"/duration="+duration, func(t *testing.T) {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "dd"), []byte("#!/bin/sh\nexec /bin/sleep 1\n"), 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
				options := &core.DiskAttackConfig{Action: core.DiskReadPayloadAction, Duration: duration, DdOptions: &[]core.DdOption{{}}}
				env := Environment{AttackUid: "duration-test", Chaos: &Server{CmdPools: map[string]*utils.CommandPools{}}}
				start := time.Now()
				var err error
				if server {
					err = ApplyDiskServerAttack(options, env)
					pool := env.Chaos.CmdPools[env.AttackUid]
					pool.Wait()
					pool.Close()
				} else {
					err = ApplyDiskAttack(options, env)
				}
				elapsed := time.Since(start)
				if duration != "" {
					if elapsed >= 800*time.Millisecond {
						t.Errorf("payload ran %s past 100ms duration", elapsed)
					}
					if err != nil {
						t.Errorf("normal duration expiry failed: %v", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if elapsed < time.Second {
						t.Errorf("payload stopped before natural completion: %s", elapsed)
					}
				}
			})
		}
	}
}

func TestDiskDurationRemainsRecoverable(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "payload")
	t.Setenv("CHAOSD_DISK_FILE", file)
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	if err := os.WriteFile(filepath.Join(dir, "dd"), []byte("#!/bin/sh\ntouch \"$CHAOSD_DISK_FILE\"\nexec /bin/sleep 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "disk.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	store := experiment.NewStore(&dbstore.DB{DB: db})
	server := NewServer(nil, store, nil, nil, nil, nil, nil, scheduler.Scheduler{})
	options := &core.DiskAttackConfig{
		Kind: core.DiskAttack, Action: core.DiskWritePayloadAction, Duration: "100ms",
		Path: file, DdOptions: &[]core.DdOption{{}},
	}
	uid, err := server.ExecuteAttack(DiskAttack, options, core.CommandMode)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := store.FindByUid(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if exp.Status != core.Success {
		t.Fatalf("status = %s", exp.Status)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal(err)
	}
	recovery := NewServer(nil, store, nil, nil, nil, nil, nil, scheduler.Scheduler{})
	if err := recovery.RecoverAttack(uid); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("payload remains after recovery: %v", err)
	}
}

func TestDiskPayloadReportsCommandFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	if err := os.WriteFile(filepath.Join(dir, "dd"), []byte("#!/bin/sh\nexit 17\n"), 0700); err != nil {
		t.Fatal(err)
	}
	options := &core.DiskAttackConfig{Action: core.DiskReadPayloadAction, Duration: "1s", DdOptions: &[]core.DdOption{{}}}
	env := Environment{AttackUid: "failure-test", Chaos: &Server{CmdPools: map[string]*utils.CommandPools{}}}
	if err := ApplyDiskAttack(options, env); err == nil {
		t.Fatal("command failure was ignored")
	}
}
