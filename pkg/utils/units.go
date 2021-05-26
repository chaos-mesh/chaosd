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

// ParseUnit parse a digit with unit such as "K" , "KiB", "KB", "c", "MiB", "MB", "M".
// If input string is a digit without unit ,
// it will be regarded as a digit with unit M(1024*1024 bytes).
func ParseUnit(s string) (uint64, error) {
	if _, err := strconv.Atoi(s); err == nil {
		s += "M"
	}
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

// DdArgBlock is command arg for dd. BlockSize is bs.Count is count.
type DdArgBlock struct {
	BlockSize string
	Count     string
}

// This func split bytes in to processNum + 1 dd arg blocks.
// Every ddArgBlock can generate one dd command.
// If bytes is bigger than processNum M ,
// bytes will be split into processNum dd commands with bs = 1M ,count = bytes/ processNum M.
// If bytes is not bigger than processNum M ,
// bytes will be split into processNum dd commands with bs = bytes / uint64(processNum) ,count = 1.
// And one ddArgBlock stand by the rest bytes will also add to the end of slice,
// even if rest bytes = 0.
func SplitBytesByProcessNum(bytes uint64, processNum uint8) ([]DdArgBlock, error) {
	if bytes == 0 {
		return []DdArgBlock{{
			BlockSize: "1M",
			Count:     "0",
		}}, nil
	}
	if processNum == 0 {
		return nil, fmt.Errorf("num must not be zero")
	}
	ddArgBlocks := make([]DdArgBlock, processNum)
	if bytes > uint64(processNum)*(1<<20) {
		count := (bytes >> 20) / uint64(processNum)
		for i := range ddArgBlocks {
			ddArgBlocks[i].Count = strconv.FormatUint(count, 10)
			ddArgBlocks[i].BlockSize = "1M"
			bytes -= count << 20
		}
	} else {
		blockSize := bytes / uint64(processNum)
		for i := range ddArgBlocks {
			ddArgBlocks[i].Count = "1"
			ddArgBlocks[i].BlockSize = strconv.FormatUint(blockSize, 10) + "c"
			bytes -= blockSize
		}
	}
	ddArgBlocks = append(ddArgBlocks, DdArgBlock{
		BlockSize: "1",
		Count:     strconv.FormatUint(bytes, 10) + "c",
	})
	return ddArgBlocks, nil
}
