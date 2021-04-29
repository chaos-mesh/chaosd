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
		want    []SizeBlock
		wantErr bool
	}{
		{
			name: "0",
			args: args{
				b:   0,
				num: 0,
			},
			want: []SizeBlock{{
				BlockSize: "1M",
				Size:      "0",
			}},
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
			want: []SizeBlock{{
				BlockSize: "524288c",
				Size:      "1",
			}, {
				BlockSize: "524288c",
				Size:      "1",
			}},
			wantErr: false,
		}, {
			name: "3",
			args: args{
				b:   1<<20 + 1,
				num: 2,
			},
			want: []SizeBlock{{
				BlockSize: "524288c",
				Size:      "1",
			}, {
				BlockSize: "524289c",
				Size:      "1",
			}},
			wantErr: false,
		}, {
			name: "4",
			args: args{
				b:   5 << 20,
				num: 2,
			},
			want: []SizeBlock{{
				BlockSize: "1M",
				Size:      "2",
			}, {
				BlockSize: "1M",
				Size:      "3",
			}},
			wantErr: false,
		}, {
			name: "5",
			args: args{
				b:   5<<20 + 1,
				num: 2,
			},
			want: []SizeBlock{{
				BlockSize: "1M",
				Size:      "2",
			}, {
				BlockSize: "3145729c",
				Size:      "1",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitByteSize(tt.args.b, tt.args.num)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitByteSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitByteSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}
