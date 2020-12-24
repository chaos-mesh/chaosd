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

package utils

import (
	"context"
	"encoding/json"

	//"fmt"
	"syscall"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaosd/pkg/core"
	"github.com/chaos-mesh/chaosd/pkg/server/chaosd"
)

func RecoverExp(expStore core.ExperimentStore, chaos *chaosd.Server, uid string) error {
	exp, err := expStore.FindByUid(context.Background(), uid)
	if err != nil {
		return err
	}

	if exp == nil {
		return errors.Errorf("experiment %s not found", uid)
	}

	if exp.Status != core.Success {
		return errors.Errorf("can not recover %s experiment", exp.Status)
	}

	switch exp.Kind {
	case core.ProcessAttack:
		pcmd := &core.ProcessCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), pcmd); err != nil {
			return err
		}

		if pcmd.Signal != int(syscall.SIGSTOP) {
			return errors.Errorf("process attack %s not support to recover", uid)
		}

		if err := chaos.RecoverProcessAttack(uid, pcmd); err != nil {
			return errors.Errorf("Recover experiment %s failed, %s", uid, err.Error())
		}
	case core.NetworkAttack:
		ncmd := &core.NetworkCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), ncmd); err != nil {
			return err
		}

		if err := chaos.RecoverNetworkAttack(uid, ncmd); err != nil {
			return errors.Errorf("Recover experiment %s failed, %s", uid, err.Error())
		}
	case core.StressAttack:
		scmd := &core.StressCommand{}
		if err := json.Unmarshal([]byte(exp.RecoverCommand), scmd); err != nil {
			return err
		}

		if err := chaos.RecoverStressAttack(uid, scmd); err != nil {
			return err
		}
	default:
		return errors.Errorf("chaos experiment kind %s not found", exp.Kind)
	}

	return nil
}
