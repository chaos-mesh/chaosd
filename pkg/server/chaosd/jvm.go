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
	"os/exec"
	"strings"
	"text/template"

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

const stressRuleTemplate = `
RULE {{.Name}}
STRESS {{.StressType}}
{{.StressValueName}} {{.StressValue}}
ENDRULE
`

const gcRuleTemplate = `
RULE {{.Name}}
GC
ENDRULE
`

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

	// submit rules
	ruleFile, err := j.generateRuleFile(attack)
	if err != nil {
		return err
	}

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, attack.Port, "l", ruleFile)
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

	if len(attack.Do) == 0 {
		switch attack.Action {
		case core.JVMLatencyAction:
			attack.Do = fmt.Sprintf("Thread.sleep(%d)", attack.LatencyDuration)
		case core.JVMExceptionAction:
			attack.Do = fmt.Sprintf("throw new %s", attack.ThrowException)
		case core.JVMReturnAction:
			attack.Do = fmt.Sprintf("return %s", attack.ReturnValue)
		case core.JVMStressAction:
			if attack.CPUCount > 0 {
				attack.StressType = "CPU"
				attack.StressValueName = "CPUCOUNT"
				attack.StressValue = fmt.Sprintf("%d", attack.CPUCount)
			} else {
				attack.StressType = "MEMORY"
				attack.StressValueName = "MEMORYTYPE"
				attack.StressValue = attack.MemoryType
			}
		}
	}
	buf := new(bytes.Buffer)
	var t *template.Template
	switch attack.Action {
	case core.JVMStressAction:
		t = template.Must(template.New("byteman rule").Parse(stressRuleTemplate))
	case core.JVMExceptionAction, core.JVMLatencyAction, core.JVMReturnAction:
		t = template.Must(template.New("byteman rule").Parse(ruleTemplate))
	case core.JVMGCAction:
		t = template.Must(template.New("byteman rule").Parse(gcRuleTemplate))
	default:
		return "", errors.Errorf("jvm action %s not supported", attack.Action)
	}
	if t == nil {
		return "", errors.Errorf("parse byeman rule template failed")
	}
	err = t.Execute(buf, attack)
	if err != nil {
		log.Error("executing template", zap.Error(err))
		return "", err
	}

	attack.RuleData = buf.String()

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
