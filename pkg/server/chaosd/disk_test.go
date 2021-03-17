package chaosd

import (
	"fmt"
	"github.com/chaos-mesh/chaosd/pkg/core"
	"os"
	"testing"
)

func TestServer_DiskFill(t *testing.T) {
	s := Server{
		exp:          nil,
		ipsetRule:    nil,
		iptablesRule: nil,
		tcRule:       nil,
		conf:         nil,
		svr:          nil,
	}

	tests := []struct {
		name    string
		fill    *core.DiskCommand
		wantErr bool
	}{
		{
			name: "0",
			fill: &core.DiskCommand{
				Action:          core.DiskFillAction,
				Size:            1024,
				Path:            "temp",
				FillByFallocate: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Create(tt.fill.Path)
			if err != nil {
				t.Errorf("unexpected err %v when creating temp file", err)
			}
			if f != nil {
				_ = f.Close()
			}
			_, err = s.DiskFill(tt.fill)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiskFill() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			stat, err := os.Stat(tt.fill.Path)
			if err != nil {
				t.Errorf("unexpected err %v when stat temp file", err)
			}
			fmt.Println(stat.Size())
		})
	}
}
