// Copyright 2021 Chaos Mesh Authors.
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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/uuid"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaosd/pkg/core"
)

const DDWritePayloadCommand = "dd if=/dev/zero of=%s bs=%s count=%s oflag=dsync"
const DDReadPayloadCommand = "dd if=%s of=/dev/null bs=%s count=%s iflag=dsync,direct,fullblock"

func (s *Server) DiskPayload(payload *core.DiskCommand) (uid string, err error) {
	uid = uuid.New().String()

	if err = s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.DiskAttack,
		Action:         payload.Action,
		RecoverCommand: payload.String(),
	}); err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), payload.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}
		if err := s.exp.Update(context.Background(), uid, core.Success, "", payload.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	switch payload.Action {
	case core.DiskWritePayloadAction:
		if payload.Path == "" {
			payload.Path = "/dev/null"
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf(DDWritePayloadCommand, payload.Path, "1M", strconv.FormatUint(payload.Size, 10)))
		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Error(string(output), zap.Error(err))
		} else {
			log.Info(string(output))
		}
		return uid, err
	case core.DiskReadPayloadAction:
		if payload.Path == "" {
			err := errors.Errorf("empty read payload path")
			log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
			return uid, err
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf(DDReadPayloadCommand, payload.Path, "1M", strconv.FormatUint(payload.Size, 10)))
		output, err := cmd.CombinedOutput()

		if err != nil {
			log.Error(string(output), zap.Error(err))
		} else {
			log.Info(string(output))
		}
		return uid, err
	default:
		err := errors.Errorf("invalid payload action")
		log.Error(fmt.Sprintf("payload action: %s", payload.Action), zap.Error(err))
		return uid, err
	}
}

const DDFillCommand = "dd if=/dev/zero of=%s bs=%s count=%s iflag=fullblock"
const DDFallocateCommand = "fallocate -l %sM %s"

func (s *Server) DiskFill(fill *core.DiskCommand) (uid string, err error) {
	uid = uuid.New().String()

	if err = s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid,
		Status:         core.Created,
		Kind:           core.DiskAttack,
		Action:         fill.Action,
		RecoverCommand: fill.String(),
	}); err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		if err != nil {
			if err := s.exp.Update(context.Background(), uid, core.Error, err.Error(), fill.String()); err != nil {
				log.Error("failed to update experiment", zap.Error(err))
			}
			return
		}
		if err := s.exp.Update(context.Background(), uid, core.Success, "", fill.String()); err != nil {
			log.Error("failed to update experiment", zap.Error(err))
		}
	}()

	if fill.Path == "" {
		tempFile, err := ioutil.TempFile("", "example")
		if err != nil {
			log.Error("unexpected err when open temp file", zap.Error(err))
			return uid, err
		}

		if tempFile != nil {
			err = tempFile.Close()
			if err != nil {
				log.Error("unexpected err when close temp file", zap.Error(err))
				return uid, err
			}
		} else {
			err := errors.Errorf("unexpected err : file get from ioutil.TempFile is nil")
			log.Error(fmt.Sprintf("payload action: %s", fill.Action), zap.Error(err))
			return uid, err
		}

		fill.Path = tempFile.Name()
		defer func() {
			err := os.Remove(fill.Path)
			if err != nil {
				log.Error(fmt.Sprintf("unexpected err when removing temp file %s", fill.Path), zap.Error(err))
			}
		}()
	}

	var cmd *exec.Cmd
	if fill.FillByFallocate {
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFallocateCommand, strconv.FormatUint(fill.Size, 10), fill.Path))
	} else {
		//1M means the block size. The bytes size dd read | write is (block size) * (size).
		cmd = exec.Command("bash", "-c", fmt.Sprintf(DDFillCommand, fill.Path, "1M", strconv.FormatUint(fill.Size, 10)))
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(string(output), zap.Error(err))
	} else {
		log.Info(string(output))
	}

	return uid, err
}
