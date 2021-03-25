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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/google/uuid"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

const ruleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
AT ENTRY
IF true
DO 
	{{.Do}};
ENDRULE
`
const bmInstallCommand = "bminstall.sh -b -Dorg.jboss.byteman.transform.all -Dorg.jboss.byteman.verbose -p %d %d"
const bmSubmitCommand = "bmsubmit.sh -p %d -%s %s"

func (s *Server) JVMPrepare(attack *core.JVMCommand) (string, error) {
	var err error
	uid := uuid.New().String()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.JVMAttack,
		Action:         attack.Action,
		RecoverCommand: attack.String(),
	}); err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), attack.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}

		// use the stressngPid as recover command, and will kill the pid when recover
		if err := s.exp.Update(context.Background(), uid, core.Success, "", attack.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	bmInstallCmd := fmt.Sprintf(bmInstallCommand, attack.Port, attack.Pid)
	cmd := exec.Command("bash", "-c", bmInstallCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return "", err
	}

	log.Info(string(output))
	return uid, err
}

func (s *Server) JVMAttack(attack *core.JVMCommand) (string, error) {
	var err error
	uid := uuid.New().String()

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.JVMAttack,
		Action:         attack.Action,
		RecoverCommand: attack.String(),
	}); err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), attack.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}

		// use the stressngPid as recover command, and will kill the pid when recover
		if err := s.exp.Update(context.Background(), uid, core.Success, "", attack.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	if len(attack.Do) == 0 {
		switch attack.Action {
		case core.JVMLatencyAction:
			attack.Do = fmt.Sprintf("Thread.sleep(%s)", attack.LatencyDuration)
		case core.JVMExceptionAction:
			attack.Do = fmt.Sprintf("throw new %s", attack.ThrowException)
		case core.JVMReturnAction:
			attack.Do = fmt.Sprintf("return %s", attack.ReturnValue)
		}
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("byteman rule").Parse(ruleTemplate))

	buf := new(bytes.Buffer)
	err = t.Execute(buf, attack)
	if err != nil {
		log.Error("executing template", zap.Error(err))
		return "", err
	}

	log.Info("byteman rule", zap.String("rule", string(buf.Bytes())))

	tmpfile, err := ioutil.TempFile("", "rule.btm")
	if err != nil {
		return "", err
	}

	log.Info("create btm file", zap.String("file", tmpfile.Name()))

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(buf.Bytes()); err != nil {
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, attack.Port, "l", tmpfile.Name())
	cmd := exec.Command("bash", "-c", bmSubmitCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return "", err
	}

	log.Info(string(output))
	return uid, nil
}

func (s *Server) RecoverJVMAttack(uid string, attack *core.JVMCommand) error {
	// Create a new template and parse the letter into it.
	t := template.Must(template.New("byteman rule").Parse(ruleTemplate))

	buf := new(bytes.Buffer)
	err := t.Execute(buf, attack)
	if err != nil {
		log.Error("executing template", zap.Error(err))
		return err
	}

	tmpfile, err := ioutil.TempFile("", "rule.btm")
	if err != nil {
		return err
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(buf.Bytes()); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	log.Info("create btm file", zap.String("file", tmpfile.Name()))

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, attack.Port, "u", tmpfile.Name())
	cmd := exec.Command("bash", "-c", bmSubmitCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}

	log.Info(string(output))

	if err := s.exp.Update(context.Background(), uid, core.Destroyed, "", attack.String()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
