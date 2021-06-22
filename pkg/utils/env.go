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

package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func SetRuntimeEnv() error {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	_, err = os.Stat(fmt.Sprintf("%s/tools", wd))
	if os.IsNotExist(err) {
		return err
	}

	path := os.Getenv("PATH")
	bytemanHome := fmt.Sprintf("%s/tools/byteman", wd)
	err = os.Setenv("BYTEMAN_HOME", bytemanHome)
	if err != nil {
		return err
	}
	err = os.Setenv("PATH", fmt.Sprintf("%s/tools:%s/bin:%s", wd, bytemanHome, path))
	if err != nil {
		return err
	}

	return nil
}
