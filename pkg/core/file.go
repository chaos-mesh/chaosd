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

package core

import (
	"encoding/json"

	"github.com/pingcap/errors"
)

type FileCommand struct {
	CommonAttackConfig

	FileName   string
	DirName    string
	Privilege  uint32
	SourceFile string
	DestFile   string
	Data       string
	Count      int
	FileMode   int
}

var _ AttackConfig = &FileCommand{}

const (
	FileCreateAction          = "create"
	FileModifyPrivilegeAction = "modify"
	FileDeleteAction          = "delete"
	FileRenameAction          = "rename"
	FileAppendAction          = "append"
)

func (n *FileCommand) Validate() error {
	if err := n.CommonAttackConfig.Validate(); err != nil {
		return err
	}
	switch n.Action {
	case FileCreateAction:
		return n.validFileCreate()
	case FileModifyPrivilegeAction:
		return n.validFileModify()
	case FileDeleteAction:
		return n.validFileDelete()
	case FileRenameAction:
		return n.validFileRename()
	case FileAppendAction:
		return n.valieFileAppend()
	default:
		return errors.Errorf("file action %s not supported", n.Action)
	}
}

func (n *FileCommand) validFileCreate() error {
	if len(n.FileName) == 0 && len(n.DirName) == 0 {
		return errors.New("one of file-name and dir-name is required")
	}

	return nil
}

func (n *FileCommand) validFileModify() error {
	if len(n.FileName) == 0 {
		return errors.New("file name is required")
	}

	if n.Privilege == 0 {
		return errors.New("file privilege is required")
	}

	return nil
}

func (n *FileCommand) validFileDelete() error {
	if len(n.FileName) == 0 && len(n.DirName) == 0 {
		return errors.New("one of file-name and dir-name is required")
	}

	return nil
}

func (n *FileCommand) validFileRename() error {
	if len(n.SourceFile) == 0 || len(n.DestFile) == 0 {
		return errors.New("both source file and destination file are required")
	}

	return nil
}

func (n *FileCommand) valieFileAppend() error {
	if len(n.FileName) == 0 {
		return errors.New("file-name is required")
	}

	if len(n.Data) == 0 {
		return errors.New("append data is required")
	}

	return nil
}

func (n *FileCommand) CompleteDefaults() {
	switch n.Action {
	case FileAppendAction:
		n.setDefaultForFileAppend()
	}
}

func (n *FileCommand) setDefaultForFileAppend() {
	if n.Count == 0 {
		n.Count = 1
	}
}

func (n FileCommand) RecoverData() string {
	data, _ := json.Marshal(n)
	return string(data)
}

func NewFileCommand() *FileCommand {
	return &FileCommand{
		CommonAttackConfig: CommonAttackConfig{
			Kind: FileAttack,
		},
	}
}
