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
	"fmt"
	"strconv"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/utils"
)

type fileAttack struct{}

var FileAttack AttackType = fileAttack{}

func (fileAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	attack := options.(*core.FileCommand)

	switch attack.Action {
	case core.FileCreateAction:
		if err = env.Chaos.createFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileModifyPrivilegeAction:
		if err = env.Chaos.modifyFilePrivilege(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileDeleteAction:
		if err = env.Chaos.deleteFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileRenameAction:
		if err = env.Chaos.renameFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileAppendAction:
		if err = env.Chaos.appendFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileReplaceAction:
		if err = env.Chaos.replaceFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *Server) createFile(attack *core.FileCommand, uid string) error {
	var cmdStr string
	if len(attack.DirName) > 0 {
		cmdStr = fmt.Sprintf("FileTool create --dir-name %s", attack.DirName)
	} else {
		cmdStr = fmt.Sprintf("FileTool create --file-name %s", attack.FileName)
	}

	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) modifyFilePrivilege(attack *core.FileCommand, uid string) error {
	// get the privilege of file and save it, used for recover
	cmdStr := "stat -c %a" + " " + attack.FileName
	output, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	fileModeStr := strings.Replace(output, "\n", "", -1)
	attack.OriginPrivilege, err = strconv.Atoi(string(fileModeStr))
	if err != nil {
		log.Error("transform string to int failed", zap.String("string", fileModeStr), zap.Error(err))
		return errors.WithStack(err)
	}

	// modify the file privilege
	cmdStr = fmt.Sprintf("FileTool modify --file-name %s --privilege %d", attack.FileName, attack.Privilege)
	_, err = utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// deleteFile will not really delete the file, just rename it
func (s *Server) deleteFile(attack *core.FileCommand, uid string) error {
	var source, dest string
	if len(attack.FileName) > 0 {
		dest = fmt.Sprintf("%s.%s", attack.FileName, uid)
		source = attack.FileName
	} else if len(attack.DirName) > 0 {
		dest = fmt.Sprintf("%s.%s", attack.DirName, uid)
		source = attack.DirName
	}

	return renameFile(source, dest)
}

func (s *Server) renameFile(attack *core.FileCommand, uid string) error {
	return renameFile(attack.SourceFile, attack.DestFile)
}

func renameFile(source, dest string) error {
	cmdStr := fmt.Sprintf("FileTool rename --old-name %s --new-name %s", source, dest)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) appendFile(attack *core.FileCommand, uid string) error {
	// first backup the file
	backupName := getBackupName(attack.FileName, uid)
	cmdStr := fmt.Sprintf("FileTool copy --file-name %s --copy-file-name %s", attack.FileName, backupName)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	cmdStr = fmt.Sprintf("FileTool append --count %d --data %s --file-name %s", attack.Count, attack.Data, attack.FileName)
	_, err = utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) replaceFile(attack *core.FileCommand, uid string) error {
	cmdStr := fmt.Sprintf("FileTool replace --file-name %s --origin-string %s --dest-string %s --line %d", attack.FileName, attack.OriginStr, attack.DestStr, attack.Line)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (fileAttack) Recover(exp core.Experiment, env Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	attack := config.(*core.FileCommand)

	switch attack.Action {
	case core.FileCreateAction:
		if err = env.Chaos.recoverCreateFile(attack); err != nil {
			return errors.WithStack(err)
		}
	case core.FileModifyPrivilegeAction:
		if err = env.Chaos.recoverModifyPrivilege(attack); err != nil {
			return errors.WithStack(err)
		}
	case core.FileDeleteAction:
		if err = env.Chaos.recoverDeleteFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	case core.FileRenameAction:
		if err = env.Chaos.recoverRenameFile(attack); err != nil {
			return errors.WithStack(err)
		}
	case core.FileAppendAction:
		if err = env.Chaos.recoverAppendFile(attack, env.AttackUid); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Server) recoverCreateFile(attack *core.FileCommand) error {
	var fileName string
	if len(attack.FileName) > 0 {
		fileName = attack.FileName
	} else if len(attack.DirName) > 0 {
		fileName = attack.DirName
	}

	cmdStr := fmt.Sprintf("FileTool delete --file-name %s", fileName)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverModifyPrivilege(attack *core.FileCommand) error {
	cmdStr := fmt.Sprintf("FileTool modify --file-name %s --privilege %d", attack.FileName, attack.OriginPrivilege)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// recoverDeleteFile just rename the backup file/dir
func (s *Server) recoverDeleteFile(attack *core.FileCommand, uid string) error {
	var backupName, sourceName string
	if len(attack.FileName) > 0 {
		backupName = getBackupName(attack.FileName, uid)
		sourceName = attack.FileName
	} else if len(attack.DirName) > 0 {
		backupName = getBackupName(attack.DirName, uid)
		sourceName = attack.DirName
	}

	err := renameFile(backupName, sourceName)
	if err != nil {
		log.Error("recover delete file/dir failed", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverRenameFile(attack *core.FileCommand) error {
	return renameFile(attack.DestFile, attack.SourceFile)
}

func (s *Server) recoverAppendFile(attack *core.FileCommand, uid string) error {
	backupName := getBackupName(attack.FileName, uid)
	cmdStr := fmt.Sprintf("FileTool rename --old-name %s --new-name %s", backupName, attack.FileName)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// getBackupName gets the backup file or directory name
func getBackupName(source string, uid string) string {
	return fmt.Sprintf("%s.%s", source, uid)
}
