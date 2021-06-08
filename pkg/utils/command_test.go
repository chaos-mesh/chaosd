package utils

import (
	"fmt"
	"testing"
)

func TestCommand_Unmarshal(t *testing.T) {
	type dd struct {
		If    string `dd:"if"`
		Of    string `dd:"oflag"`
		Iflag string `dd:"iflag"`
	}
	dc := Command{Name: "dd"}
	tests := []struct {
		name string
		d    dd
	}{
		{
			name: "0",
			d: dd{
				"/dev/zero",
				"i,2,3",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := dc.Unmarshal(tt.d)
			fmt.Println(cmd.String())
		})
	}
}
