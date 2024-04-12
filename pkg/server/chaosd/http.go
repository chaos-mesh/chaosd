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
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

type attackHTTP struct{}

var HTTPAttack AttackType = attackHTTP{}

func (attackHTTP) Attack(options core.AttackConfig, _ Environment) error {
	var attackConf *core.HTTPAttackConfig
	var ok bool
	if attackConf, ok = options.(*core.HTTPAttackConfig); !ok {
		return errors.New("AttackConfig -> *HTTPAttackConfig meet error")
	}

	if attackConf.Action == core.HTTPRequestAction {
		return attackHTTPRequest(attackConf)
	}

	cmd := exec.Command("tproxy", "-i", "-vv")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "create stdin pipe")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "create stdout pipe")
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "start command `%s`", cmd.String())
	}

	config, err := json.Marshal(&attackConf.Config)
	attackConf.Logger.Info(string(config))
	if err != nil {
		return errors.Wrap(err, "applying HTTP attack")
	}

	req, err := http.NewRequest(http.MethodPut, "/", bytes.NewReader(config))
	if err != nil {
		return errors.Wrap(err, "create http request")
	}

	err = req.Write(stdin)
	if err != nil {
		return errors.Wrap(err, "cannot request tproxy")
	}

	resp, err := http.ReadResponse(bufio.NewReader(stdout), req)
	if err != nil {
		return errors.Wrap(err, "cannot read response")
	}
	if resp.StatusCode != http.StatusOK {
		by, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "cannot read err resp body, %s", resp.Status)
		}
		return errors.Errorf("%s: %s", resp.Status, string(by))
	}

	by, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "cannot read resp body")
	}
	attackConf.Logger.Info(string(by))

	attackConf.ProxyPID = cmd.Process.Pid
	// In linux, a child process will become orphan process when a parent process dies.
	// But Golang runtime maintains a finalizer for a child process.
	// Release() will clear the finalizer for chaos-tproxy here.
	err = cmd.Process.Release()
	if err != nil {
		return errors.Wrapf(err, "Fatal error : release process fail, please clear PID: %d", attackConf.ProxyPID)
	}
	return nil
}

func (attackHTTP) Recover(exp core.Experiment, _ Environment) error {
	config, err := exp.GetRequestCommand()
	if err != nil {
		return err
	}
	attack, ok := config.(*core.HTTPAttackConfig)
	if !ok {
		return errors.Errorf("AttackConfig -> *HTTPAttackConfig meet error")
	}

	proc, err := process.NewProcess(int32(attack.ProxyPID))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return err
	}

	procName, err := proc.Name()
	if err != nil {
		return errors.Wrapf(err, "unexpected error when proc.Name. process pid: %d", proc.Pid)
	}

	if !strings.Contains(procName, "tproxy") {
		attack.Logger.Info("the process %s:%d is not chaos-tproxy, please check and clear it manually\n", procName, attack.ProxyPID)
		return nil
	}

	if err := proc.Terminate(); err != nil {
		attack.Logger.Info("the chaos-tproxy process kill failed with error: %s\n", err.Error())
		return nil
	}
	return nil
}

func attackHTTPRequest(attackConf *core.HTTPAttackConfig) error {
	if attackConf.EnableConnPool {
		var HTTPTransport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       60 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
			MaxIdleConnsPerHost:   100,
		}

		client := &http.Client{
			Transport: HTTPTransport,
		}
		for i := 0; i < attackConf.Count; i++ {
			req, err := http.NewRequest(http.MethodGet, attackConf.URL, nil)
			if err != nil {
				return errors.Wrap(err, "create HTTP request")
			}
			resp, err := client.Do(req)
			if err != nil {
				return errors.Wrap(err, "HTTP request")
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "read response body")
			}
			attackConf.Logger.Info("response body: " + string(data))
		}
	} else {
		for i := 0; i < attackConf.Count; i++ {
			resp, err := http.Get(attackConf.URL)
			if err != nil {
				return errors.Wrap(err, "HTTP request")
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "read response body")
			}
			attackConf.Logger.Info("response body: " + string(data))
		}
	}

	return nil
}
