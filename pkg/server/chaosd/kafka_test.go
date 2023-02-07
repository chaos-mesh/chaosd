// Copyright 2022 Chaos Mesh Authors.
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
	"io/fs"
	"os"
	"testing"
	"time"

	perr "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type fakeFileInfo struct {
	mode os.FileMode
}

func (i *fakeFileInfo) Name() string {
	panic("unimplemented")
}

func (i *fakeFileInfo) Size() int64 {
	panic("unimplemented")
}

func (i *fakeFileInfo) Mode() os.FileMode {
	return i.mode
}

func (i *fakeFileInfo) ModTime() time.Time {
	panic("unimplemented")
}

func (i *fakeFileInfo) IsDir() bool {
	panic("unimplemented")
}

func (i *fakeFileInfo) Sys() interface{} {
	panic("unimplemented")
}

type fakeFs map[string]*fakeFileInfo

func newFakeFs() fakeFs {
	return make(fakeFs)
}

func (f fakeFs) stat(path string) (fs.FileInfo, error) {
	info, ok := f[path]
	if !ok {
		return nil, perr.Errorf("fail to stat: %s", path)
	}
	return info, nil
}

func (f fakeFs) chmod(path string, mode os.FileMode) error {
	info, ok := f[path]
	if !ok {
		return perr.Errorf("fail to stat: %s", path)
	}
	info.mode = mode
	return nil
}

func (f fakeFs) clone() fakeFs {
	bak := newFakeFs()
	for p, i := range f {
		bak[p] = &fakeFileInfo{
			mode: i.mode,
		}
	}
	return bak
}

func (f fakeFs) equal(other fakeFs) bool {
	if len(f) != len(other) {
		return false
	}
	for p, info := range f {
		otherInfo, ok := other[p]
		if !ok || info.mode != otherInfo.mode {
			return false
		}
	}
	return true
}

func (f fakeFs) nonreadable() fakeFs {
	bak := f.clone()
	for _, i := range bak {
		i.mode &= ^os.FileMode(0444)
	}
	return bak
}

func (f fakeFs) nonwritable() fakeFs {
	bak := f.clone()
	for _, i := range bak {
		i.mode &= ^os.FileMode(0222)
	}
	return bak
}

func TestAttackIO(t *testing.T) {
	originFs := fakeFs{
		"/a": &fakeFileInfo{
			mode: os.FileMode(0777),
		},
		"/b": &fakeFileInfo{
			mode: os.FileMode(0700),
		},
		"/c": &fakeFileInfo{
			mode: os.FileMode(0070),
		},
		"/d": &fakeFileInfo{
			mode: os.FileMode(0007),
		},
		"/e": &fakeFileInfo{
			mode: os.FileMode(0123),
		},
		"/f": &fakeFileInfo{
			mode: os.FileMode(0124),
		},
		"/g": &fakeFileInfo{
			mode: os.FileMode(0247),
		},
	}

	assert.True(t, originFs.equal(originFs))
	assert.False(t, originFs.equal(originFs.nonreadable()))
	bakFs := originFs.clone()
	assert.True(t, originFs.equal(bakFs))

	attack := core.NewKafkaCommand()
	attack.NonReadable = true
	// make only "/a" non-readable
	err := attackIOPath(attack, "/a", bakFs.stat, bakFs.chmod)
	assert.Nil(t, err)
	assert.Equal(t, bakFs["/a"].Mode(), originFs.nonreadable()["/a"].Mode())

	// recover
	err = recoverIOPath("/a", uint32(originFs["/a"].Mode()), bakFs.chmod)
	assert.Nil(t, err)
	assert.True(t, bakFs.equal(originFs))

	// make all path non-readable
	for p := range bakFs {
		err := attackIOPath(attack, p, bakFs.stat, bakFs.chmod)
		assert.Nil(t, err)
	}
	assert.True(t, bakFs.equal(originFs.nonreadable()))

	// recover
	for p, mode := range attack.OriginModeOfFiles {
		err := recoverIOPath(p, mode, bakFs.chmod)
		assert.Nil(t, err)
	}
	assert.True(t, bakFs.equal(originFs))

	// make all path non-readable and non-writable
	attack.NonWritable = true
	for p := range bakFs {
		err := attackIOPath(attack, p, bakFs.stat, bakFs.chmod)
		assert.Nil(t, err)
	}
	assert.True(t, bakFs.equal(originFs.nonreadable().nonwritable()))

	// recover
	for p, mode := range attack.OriginModeOfFiles {
		err := recoverIOPath(p, mode, bakFs.chmod)
		assert.Nil(t, err)
	}
	assert.True(t, bakFs.equal(originFs))
}
