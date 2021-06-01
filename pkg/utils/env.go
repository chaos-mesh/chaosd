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
	"fmt"
	"os"
	"path/filepath"
)

func SetRuntimeEnv() {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return
	}

	_, err = os.Stat(fmt.Sprintf("%s/tools", wd))
	if os.IsNotExist(err) {
		return
	}

	path := os.Getenv("PATH")
	bytemanHome := fmt.Sprintf("%s/tools/byteman", wd)
	os.Setenv("BYTEMAN_HOME", bytemanHome)
	os.Setenv("PATH", fmt.Sprintf("%s/tools:%s/bin:%s", wd, bytemanHome, path))

	//fmt.Println("BYTEMAN_HOME:", os.Getenv("BYTEMAN_HOME"))
	//fmt.Println("PATH:", os.Getenv("PATH"))
}
