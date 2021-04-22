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
	"syscall"
)

// GetDiskTotalSize returns the total bytes in disk
func GetDiskTotalSize(path string) (total uint64, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return 0, err
	}
	reservedBlocks := s.Bfree - s.Bavail
	total = uint64(s.Frsize) * (s.Blocks - reservedBlocks)
	return total, nil
}
