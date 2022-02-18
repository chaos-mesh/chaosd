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
	"io/ioutil"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

// CreateTempFile will create a temp file in current directory.
func CreateTempFile(path string) (string, error) {
	tempFile, err := ioutil.TempFile(path, "example")
	if err != nil {
		log.Error("unexpected err when open temp file", zap.Error(err))
		return "", err
	}

	if tempFile != nil {
		err = tempFile.Close()
		if err != nil {
			log.Error("unexpected err when close temp file", zap.Error(err))
			return "", err
		}
	} else {
		err := errors.Errorf("unexpected err : file get from ioutil.TempFile is nil")
		return "", err
	}
	return tempFile.Name(), nil
}
