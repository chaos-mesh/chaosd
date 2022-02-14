// Copyright 2020 Chaos Mesh Authors.
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
	"os"
	"os/exec"
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
	}

	return nil
}

func (s *Server) createFile(attack *core.FileCommand, uid string) error {
	cmdStr := fmt.Sprintf("FileTool create --dest-dir \"%s\" --dir-name \"%s\" --file-name \"%s\"", attack.DestDir, attack.DirName, attack.FileName)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) modifyFilePrivilege(attack *core.FileCommand, uid string) error {
	cmdStr := "stat -c %a" + " " + attack.FileName
	output, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	fileModeStr := strings.Replace(output, "\n", "", -1)
	attack.FileMode, err = strconv.Atoi(string(fileModeStr))
	if err != nil {
		log.Error("transform string to int failed", zap.String("string", fileModeStr), zap.Error(err))
		return errors.WithStack(err)
	}

	cmdStr = fmt.Sprintf("FileTool modify --file-name %s --privilege %d", attack.FileName, attack.Privilege)
	_, err = utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// deleteFile will not really delete the file, just rename it
func (s *Server) deleteFile(attack *core.FileCommand, uid string) error {
	var sourceFile, destFile string
	if len(attack.FileName) > 0 {
		destFile = attack.DestDir + attack.FileName + "." + uid
		sourceFile = fmt.Sprintf("%s/%s", attack.DestDir, attack.FileName)
	} else if len(attack.DirName) > 0 {
		destFile = attack.DestDir + attack.DirName + "." + uid
		sourceFile = fmt.Sprintf("%s/%s", attack.DestDir, attack.DirName)
	}

	return renameFile(sourceFile, destFile)
}

func (s *Server) renameFile(attack *core.FileCommand, uid string) error {
	return renameFile(attack.SourceFile, attack.DstFile)
}

func renameFile(sourceFile, dstFile string) error {
	cmdStr := fmt.Sprintf("FileTool rename --old-name %s --new-name %s", sourceFile, dstFile)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

//while the input content has many lines, "\n" is the line break
func (s *Server) appendFile(attack *core.FileCommand, uid string) error {
	cmdStr := fmt.Sprintf("FileTool append --count %d --data %s --file-name", attack.Count, attack.Data, attack.FileName)
	_, err := utils.ExecuteCmd(cmdStr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func deleteTestFile(file string) error {
	cmdStr := fmt.Sprintf("rm -rf %s", file)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print("delete test file error")
		log.Error(string(output), zap.Error(err))
		return errors.WithStack(err)
	}
	log.Info(string(output))

	return nil
}

func fileExist(fileName string) bool {
	_, err := os.Lstat(fileName)
	return !os.IsNotExist(err)
}

func fileEmpty(fileName string) bool {
	file, err := os.Stat(fileName)
	if err != nil {
		log.Error("get file is empty err", zap.Error(err))
	}
	if file.Size() == 0 {
		return true
	}
	return false
}

func generateFile(s string) (*os.File, error) {
	fileName := "test.dat"
	dstFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
		return dstFile, err
	}
	defer dstFile.Close()
	strNew := strings.Replace(s, `\n`, "\n", -1)
	dstFile.WriteString(strNew)
	return dstFile, nil
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
		if err = env.Chaos.recoverAppendFile(attack); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Server) recoverCreateFile(attack *core.FileCommand) error {

	var err error
	if len(attack.FileName) > 0 {
		err = os.Remove(attack.DestDir + attack.FileName)
	} else if len(attack.DirName) > 0 {
		err = os.RemoveAll(attack.DestDir + attack.DirName)
	}

	if err != nil {
		log.Error("delete file/dir faild", zap.Error(err))
		return errors.WithStack(err)
	}
	return nil
}

func (s *Server) recoverModifyPrivilege(attack *core.FileCommand) error {

	cmdStr := fmt.Sprintf("chmod %d %s", attack.FileMode, attack.FileName)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverDeleteFile(attack *core.FileCommand, uid string) error {
	var err error
	if len(attack.FileName) > 0 {
		backFile := attack.DestDir + attack.FileName + "." + uid
		err = os.Rename(backFile, attack.DestDir+attack.FileName)
	} else if len(attack.DirName) > 0 {
		backDir := attack.DestDir + attack.DirName + "." + uid
		err = os.Rename(backDir, attack.DestDir+attack.DirName)
	}

	if err != nil {
		log.Error("recover delete file/dir failed", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverRenameFile(attack *core.FileCommand) error {
	err := os.Rename(attack.DstFile, attack.SourceFile)

	if err != nil {
		log.Error("recover rename file/dir faild", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) recoverAppendFile(attack *core.FileCommand) error {
	// after attack, delete the generated file
	if fileExist("test.dat") {
		if err := deleteTestFile("test.dat"); err != nil {
			return errors.WithStack(err)
		}
	}

	// count the number of rows inserted
	linesByInput := attack.Count * core.GetFileNumber(attack.FileName)

	// delete linesByInput rows starting with the inserted row
	c := fmt.Sprintf("%d,%dd", attack.LineNo+1, attack.LineNo+linesByInput)
	cmdStr := fmt.Sprintf("sed -i '%s' %s", c, attack.FileName)

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}
