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
	"strconv"

	"github.com/alecthomas/units"
)

var (
	// See https://en.wikipedia.org/wiki/Binary_prefix
	shortBinaryUnitMap = units.MakeUnitMap("", "c", 1024)
	binaryUnitMap      = units.MakeUnitMap("iB", "c", 1024)
	decimalUnitMap     = units.MakeUnitMap("B", "c", 1000)
)

func ParseUnit(s string) (uint64, error) {
	if n, err := units.ParseUnit(s, shortBinaryUnitMap); err == nil {
		return uint64(n), nil
	}

	if n, err := units.ParseUnit(s, binaryUnitMap); err == nil {
		return uint64(n), nil
	}

	if n, err := units.ParseUnit(s, decimalUnitMap); err == nil {
		return uint64(n), nil
	}
	return 0, fmt.Errorf("units: unknown unit %s", s)
}

type SizeBlock struct {
	BlockSize string
	Size      string
}

func SplitByteSize(b uint64, num uint8) []SizeBlock {
	if b == 0 {
		return []SizeBlock{{
			BlockSize: "1M",
			Size:      "0",
		}}
	}
	sizeBlocks := make([]SizeBlock, num)
	if b > uint64(num)*(1<<20) {
		splitSize := (b >> 20) / uint64(num)
		for i := range sizeBlocks {
			if i == len(sizeBlocks)-1 {
				if (b >> 20 << 20) == b {
					sizeBlocks[i].Size = strconv.FormatUint((b>>20)%uint64(num), 10)
					sizeBlocks[i].BlockSize = "1M"
				} else {
					sizeBlocks[i].Size = "1"
					sizeBlocks[i].BlockSize = strconv.FormatUint(splitSize<<20+b%(splitSize<<20), 10) + "c"
				}
			} else {
				sizeBlocks[i].Size = strconv.FormatUint(splitSize, 10)
				sizeBlocks[i].BlockSize = "1M"
				b -= splitSize
			}
		}
	} else {
		for i := range sizeBlocks {
			if i != len(sizeBlocks)-1 {
				sizeBlocks[i].Size = "1"
				sizeBlocks[i].BlockSize = strconv.FormatUint(b/uint64(num), 10) + "c"
			} else {
				sizeBlocks[i].Size = "1"
				sizeBlocks[i].BlockSize = strconv.FormatUint(b/uint64(num)+b%uint64(num), 10) + "c"
			}
		}
	}
	return sizeBlocks
}
