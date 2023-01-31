// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.
package core

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_initSize(t *testing.T) {
	opt := DiskOption{
		CommonAttackConfig: CommonAttackConfig{
			Action: DiskFillAction,
		},
		Size: "1024M",
	}
	byteSize, err := initSize(&opt)
	assert.NoError(t, err)
	assert.EqualValues(t, 1024<<20, byteSize)

	opt.Percent = "99%"
	opt.Size = ""
	byteSize, err = initSize(&opt)
	assert.NoError(t, err)
	t.Logf("percent %s with bytesize %sGB\n", opt.Percent, strconv.Itoa(int(byteSize>>30)))

	opt.Percent = ""
	opt.Size = ""
	_, err = initSize(&opt)
	assert.Error(t, err)
}

func Test_initPath(t *testing.T) {
	opt := DiskOption{
		CommonAttackConfig: CommonAttackConfig{
			Action: DiskFillAction,
		},
		Path: "/1/12/1/2/1/21",
	}
	_, err := initPath(&opt)
	assert.Error(t, err)
}
