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

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModifyFilePrivilegeWithSpaceInName(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "my file")
	if err := os.WriteFile(fileName, nil, 0o640); err != nil {
		t.Fatal(err)
	}

	if err := modifyFilePrivilege(fileName, 600); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode %o, want 600", info.Mode().Perm())
	}
}

func TestReplaceFileWithSpacesAndQuotes(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "my file")
	if err := os.WriteFile(fileName, []byte("foo bar baz\nfoo bar baz\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(fileName, "bar baz", "it's", 2); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if want := "foo bar baz\nfoo it's\n"; string(data) != want {
		t.Fatalf("file content %q, want %q", data, want)
	}
}
