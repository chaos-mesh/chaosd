// Copyright 2020 Chaos Mesh Authors.
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

package utils

import (
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/pingcap/log"
)

// GetProgramPath returns the path of the program
func GetProgramPath() string {
	dir, err := exec.LookPath(os.Args[0])
	if err != nil {
		log.Fatal("can not get the process path", zap.Error(err))
	}
	if p, err := os.Readlink(dir); err == nil {
		dir = p
	}
	proPath, err := filepath.Abs(filepath.Dir(dir))
	if err != nil {
		log.Fatal("can not get the full process path", zap.Error(err))
	}
	return proPath
}
