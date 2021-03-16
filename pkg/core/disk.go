package core

import "encoding/json"

const (
	DiskFillAction         = "fill"
	DiskWritePayloadAction = "write-payload"
	DiskReadPayloadAction  = "read-payload"
)

type DiskCommand struct {
	Action string
	Size   uint64
	Path   string
}

func (d *DiskCommand) Validate() error {
	return nil
}

func (d *DiskCommand) String() string {
	data, _ := json.Marshal(d)

	return string(data)
}
