// Copyright 2023 Chaos Mesh Authors.
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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

func Test_diskAttack_Attack(t *testing.T) {
	opt := core.DiskOption{
		CommonAttackConfig: core.CommonAttackConfig{
			Action: core.DiskFillAction,
		},
		Size:              "10M",
		Path:              "./a",
		PayloadProcessNum: 1,
	}
	env := Environment{
		AttackUid: "a",
		Chaos: &Server{
			CmdPools: make(map[string]*utils.CommandPools),
		},
	}
	conf, err := opt.PreProcess()
	assert.NoError(t, err)
	err = DiskAttack.Attack(conf, env)
	assert.NoError(t, err)

	f, err := os.Open("./a")
	assert.NoError(t, err)
	fi, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), fi.Size()>>20)
	err = os.Remove("./a")
	assert.NoError(t, err)

	opt.Action = core.DiskWritePayloadAction
	opt.PayloadProcessNum = 4
	wConf, err := opt.PreProcess()
	assert.NoError(t, err)
	err = DiskAttack.Attack(wConf, env)
	assert.NoError(t, err)

	f, err = os.Open("./a")
	assert.NoError(t, err)
	fi, err = f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, fi.Size()>>20, int64(2))
	err = os.Remove("./a")
	assert.NoError(t, err)

	opt.Action = core.DiskReadPayloadAction
	opt.PayloadProcessNum = 4
	opt.Path = "./"
	_, err = opt.PreProcess()
	assert.Error(t, err)
}
