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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type jvmAttack struct{}

var JVMAttack AttackType = jvmAttack{}

const bmInstallCommand = "bminstall.sh -b -Dorg.jboss.byteman.transform.all -Dorg.jboss.byteman.verbose -p %d %d"
const bmSubmitCommand = "bmsubmit.sh -p %d -%s %s"

func (j jvmAttack) Attack(options core.AttackConfig, env Environment) (err error) {
	// install agent
	attack := options.(*core.JVMCommand)
	bmInstallCmd := fmt.Sprintf(bmInstallCommand, attack.Port, attack.Pid)
	cmd := exec.Command("bash", "-c", bmInstallCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// this error will occured when install agent more than once, and will ignore this error and continue to submit rule
		errMsg1 := "Agent JAR loaded but agent failed to initialize"

		// these two errors will occured when java version less or euqal to 1.8, and don't know why
		// but it can install agent success even with this error, so just ignore it now.
		// TODO: Investigate the cause of these two error
		errMsg2 := "Provider sun.tools.attach.LinuxAttachProvider not found"
		errMsg3 := "install java.io.IOException: Non-numeric value found"
		if !strings.Contains(string(output), errMsg1) && !strings.Contains(string(output), errMsg2) &&
			!strings.Contains(string(output), errMsg3) {
			log.Error(string(output), zap.Error(err))
			return err
		}
		log.Debug(string(output), zap.Error(err))
	}

	// submit helper jar
	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, attack.Port, "b", fmt.Sprintf("%s/lib/byteman-helper.jar", os.Getenv("BYTEMAN_HOME")))
	cmd = exec.Command("bash", "-c", bmSubmitCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}
	if len(output) > 0 {
		log.Info("submit helper", zap.String("output", string(output)))
	}

	// submit rules
	ruleFile, err := j.generateRuleFile(attack)
	if err != nil {
		return err
	}

	bmSubmitCmd = fmt.Sprintf(bmSubmitCommand, attack.Port, "l", ruleFile)
	cmd = exec.Command("bash", "-c", bmSubmitCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}

	if len(output) > 0 {
		log.Info("submit rules", zap.String("output", string(output)))
	}

	return nil
}

func (j jvmAttack) generateRuleFile(attack *core.JVMCommand) (string, error) {
	var err error
	if len(attack.RuleData) > 0 {
		filename, err := writeDataIntoFile(attack.RuleData, "rule.btm")
		if err != nil {
			return "", err
		}
		log.Info("byteman rule", zap.String("rule", string(attack.RuleData)), zap.String("file", filename))

		return filename, nil
	}

	if len(attack.RuleFile) > 0 {
		data, err := ioutil.ReadFile(attack.RuleFile)
		if err != nil {
			return "", err
		}
		attack.RuleData = string(data)
		log.Info("rule file data:" + attack.RuleData)

		return attack.RuleFile, nil
	}

	attack.RuleData, err = generateRuleData(attack)
	if err != nil {
		return "", err
	}

	filename, err := writeDataIntoFile(attack.RuleData, "rule.btm")
	if err != nil {
		return "", err
	}
	log.Info("byteman rule", zap.String("rule", attack.RuleData), zap.String("file", filename))

	return filename, nil
}

func (j jvmAttack) Recover(exp core.Experiment, env Environment) error {
	attack := &core.JVMCommand{}
	if err := json.Unmarshal([]byte(exp.RecoverCommand), attack); err != nil {
		return err
	}

	filename, err := writeDataIntoFile(attack.RuleData, "rule.btm")
	if err != nil {
		return err
	}
	log.Info("create btm file", zap.String("file", filename))

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, attack.Port, "u", filename)
	cmd := exec.Command("bash", "-c", bmSubmitCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output), zap.Error(err))
		return err
	}

	log.Info(string(output))

	return nil
}

func generateRuleData(attack *core.JVMCommand) (string, error) {
	bytemanTemplateSpec := core.BytemanTemplateSpec{
		Name:   attack.Name,
		Class:  attack.Class,
		Method: attack.Method,
	}

	var mysqlException string
	switch attack.Action {
	case core.JVMLatencyAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("Thread.sleep(%d)", attack.LatencyDuration)
	case core.JVMExceptionAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("throw new %s", attack.ThrowException)
	case core.JVMReturnAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("return %s", attack.ReturnValue)
	case core.JVMStressAction:
		bytemanTemplateSpec.Helper = core.StressHelper
		bytemanTemplateSpec.Class = core.TriggerClass
		bytemanTemplateSpec.Method = core.TriggerMethod
		if attack.CPUCount > 0 {
			bytemanTemplateSpec.Do = fmt.Sprintf("injectCPUStress(\"%s\", %d)", attack.Name, attack.CPUCount)
		} else {
			bytemanTemplateSpec.Do = fmt.Sprintf("injectMemStress(\"%s\", %s)", attack.Name, attack.MemoryType)
		}
	case core.JVMGCAction:
		bytemanTemplateSpec.Helper = core.GCHelper
		bytemanTemplateSpec.Class = core.TriggerClass
		bytemanTemplateSpec.Method = core.TriggerMethod
		bytemanTemplateSpec.Do = "gc()"
	case core.JVMMySQLAction:
		bytemanTemplateSpec.Helper = core.SQLHelper
		// the first parameter of matchDBTable is the database which the SQL execute in, because the SQL may not contain database, for example: select * from t1;
		// can't get the database information now, so use a "" instead
		// TODO: get the database information and fill it in matchDBTable function
		bytemanTemplateSpec.Bind = fmt.Sprintf("flag:boolean=matchDBTable(\"\", $2, \"%s\", \"%s\", \"%s\")", attack.Database, attack.Table, attack.SQLType)
		bytemanTemplateSpec.Condition = "flag"
		if attack.MySQLConnectorVersion == "5" {
			bytemanTemplateSpec.Class = core.MySQL5InjectClass
			bytemanTemplateSpec.Method = core.MySQL5InjectMethod
			mysqlException = core.MySQL5Exception
		} else if attack.MySQLConnectorVersion == "8" {
			bytemanTemplateSpec.Class = core.MySQL8InjectClass
			bytemanTemplateSpec.Method = core.MySQL8InjectMethod
			mysqlException = core.MySQL8Exception
		} else {
			return "", errors.Errorf("mysql connector version %s is not supported", attack.MySQLConnectorVersion)
		}

		if len(attack.ThrowException) > 0 {
			exception := fmt.Sprintf(mysqlException, attack.ThrowException)
			bytemanTemplateSpec.Do = fmt.Sprintf("throw new %s", exception)
		} else if attack.LatencyDuration > 0 {
			bytemanTemplateSpec.Do = fmt.Sprintf("Thread.sleep(%d)", attack.LatencyDuration)
		}
	}

	buf := new(bytes.Buffer)
	var t *template.Template
	switch attack.Action {
	case core.JVMStressAction, core.JVMGCAction, core.JVMMySQLAction:
		t = template.Must(template.New("byteman rule").Parse(core.CompleteRuleTemplate))
	case core.JVMExceptionAction, core.JVMLatencyAction, core.JVMReturnAction:
		t = template.Must(template.New("byteman rule").Parse(core.SimpleRuleTemplate))
	default:
		return "", errors.Errorf("jvm action %s not supported", attack.Action)
	}
	if t == nil {
		return "", errors.Errorf("parse byeman rule template failed")
	}
	err := t.Execute(buf, bytemanTemplateSpec)
	if err != nil {
		log.Error("executing template", zap.Error(err))
		return "", err
	}

	return buf.String(), nil
}

func writeDataIntoFile(data string, filename string) (string, error) {
	tmpfile, err := ioutil.TempFile("", filename)
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.WriteString(data); err != nil {
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), err
}
