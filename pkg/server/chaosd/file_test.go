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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

// fakeFileTool puts a FileTool on PATH that records each invocation's arguments, one per line.
func fakeFileTool(t *testing.T) string {
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := "#!/bin/sh\nfor arg in \"$@\"; do printf '%s\\n' \"$arg\" >> '" + argsFile + "'; done\n"
	if err := os.WriteFile(filepath.Join(dir, "FileTool"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return argsFile
}

func TestFileAttackPassesArgumentsUnchanged(t *testing.T) {
	// modify reads the current privilege of the file before calling FileTool.
	privilegeFile := filepath.Join(t.TempDir(), "my file")
	if err := os.WriteFile(privilegeFile, nil, 0o640); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name    string
		command *core.FileCommand
		want    []string
	}{
		{
			name:    "create",
			command: &core.FileCommand{CommonAttackConfig: core.CommonAttackConfig{Action: core.FileCreateAction}, FileName: "/tmp/my dir/new file"},
			want:    []string{"create", "--file-name", "/tmp/my dir/new file"},
		},
		{
			name:    "modify",
			command: &core.FileCommand{CommonAttackConfig: core.CommonAttackConfig{Action: core.FileModifyPrivilegeAction}, FileName: privilegeFile, Privilege: 600},
			want:    []string{"modify", "--file-name", privilegeFile, "--privilege", "600"},
		},
		{
			name:    "rename",
			command: &core.FileCommand{CommonAttackConfig: core.CommonAttackConfig{Action: core.FileRenameAction}, SourceFile: "/tmp/a b", DestFile: "/tmp/c;d"},
			want:    []string{"rename", "--old-name", "/tmp/a b", "--new-name", "/tmp/c;d"},
		},
		{
			name:    "append",
			command: &core.FileCommand{CommonAttackConfig: core.CommonAttackConfig{Action: core.FileAppendAction}, FileName: "/tmp/a b", Data: "hello 'chaos' $HOME", Count: 2},
			want: []string{
				"copy", "--file-name", "/tmp/a b", "--copy-file-name", "/tmp/a b.uid",
				"append", "--count", "2", "--data", "hello 'chaos' $HOME", "--file-name", "/tmp/a b",
			},
		},
		{
			name:    "replace",
			command: &core.FileCommand{CommonAttackConfig: core.CommonAttackConfig{Action: core.FileReplaceAction}, FileName: "/tmp/a b", OriginStr: "foo bar", DestStr: "it's", Line: 3},
			want:    []string{"replace", "--file-name", "/tmp/a b", "--origin-string", "foo bar", "--dest-string", "it's", "--line", "3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			argsFile := fakeFileTool(t)
			if err := FileAttack.Attack(tc.command, Environment{AttackUid: "uid", Chaos: &Server{}}); err != nil {
				t.Fatal(err)
			}

			data, err := os.ReadFile(argsFile)
			if err != nil {
				t.Fatal(err)
			}
			got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("FileTool arguments %q, want %q", got, tc.want)
			}
		})
	}
}
