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
	"reflect"
	"testing"
)

func TestParseUnit(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    uint64
		wantErr bool
	}{
		{
			name:    "0",
			args:    "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "1",
			args:    "1",
			want:    1 << 20,
			wantErr: false,
		},
		{
			name:    "1MS",
			args:    "1MS",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUnit(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseUnit() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitByteSize(t *testing.T) {
	type args struct {
		b   uint64
		num uint8
	}
	tests := []struct {
		name    string
		args    args
		want    []DdArgBlock
		wantErr bool
	}{
		{
			name: "0",
			args: args{
				b:   0,
				num: 0,
			},
			want: []DdArgBlock{
				{
					BlockSize: "1M",
					Count:     "0",
				},
				{
					BlockSize: "",
					Count:     "",
				},
			},
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				b:   1 << 20,
				num: 0,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "2",
			args: args{
				b:   1 << 20,
				num: 2,
			},
			want: []DdArgBlock{{
				BlockSize: "524288c",
				Count:     "1",
			}, {
				BlockSize: "524288c",
				Count:     "1",
			}, {
				BlockSize: "1",
				Count:     "0c",
			}},
			wantErr: false,
		}, {
			name: "3",
			args: args{
				b:   1<<20 + 1,
				num: 2,
			},
			want: []DdArgBlock{{
				BlockSize: "524288c",
				Count:     "1",
			}, {
				BlockSize: "524288c",
				Count:     "1",
			}, {
				BlockSize: "1",
				Count:     "1c",
			}},
			wantErr: false,
		}, {
			name: "4",
			args: args{
				b:   5 << 20,
				num: 2,
			},
			want: []DdArgBlock{{
				BlockSize: "1M",
				Count:     "2",
			}, {
				BlockSize: "1M",
				Count:     "2",
			}, {
				BlockSize: "1",
				Count:     "1048576c",
			}},
			wantErr: false,
		}, {
			name: "5",
			args: args{
				b:   5<<20 + 1,
				num: 2,
			},
			want: []DdArgBlock{{
				BlockSize: "1M",
				Count:     "2",
			}, {
				BlockSize: "1M",
				Count:     "2",
			}, {
				BlockSize: "1",
				Count:     "1048577c",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitBytesByProcessNum(tt.args.b, tt.args.num)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitBytesByProcessNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitBytesByProcessNum() got = %v, want %v", got, tt.want)
			}
		})
	}
}
