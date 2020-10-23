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
	"context"

	"go.uber.org/zap"

	"github.com/pingcap/log"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/server/pb"
	"github.com/chaos-mesh/chaos-daemon/pkg/time"
)

func (s *Server) SetContainerTime(ctx context.Context, req *pb.TimeRequest) error {
	pid, err := s.criCli.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error("failed to get pid", zap.Any("request", req))
		return errors.WithStack(err)
	}

	childPids, err := GetChildProcesses(pid)
	if err != nil {
		log.Error("failed to get child processes", zap.Error(err))
	}
	allPids := append(childPids, pid)
	log.Info("all related processes found", zap.Uint32s("pids", allPids))

	for _, pid := range allPids {
		err = time.ModifyTime(int(pid), req.Sec, req.Nsec, req.ClkIdsMask)
		if err != nil {
			log.Error("failed to modify time", zap.Uint32("pid", pid), zap.Error(err))
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *Server) RecoverContainerTime(ctx context.Context, req *pb.TimeRequest) error {
	pid, err := s.criCli.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error("failed to get pid", zap.Error(err))
		return errors.WithStack(err)
	}

	childPids, err := GetChildProcesses(pid)
	if err != nil {
		log.Error("failed to get child processes", zap.Error(err))
	}
	allPids := append(childPids, pid)
	log.Info("get all related process pids", zap.Uint32s("pids", allPids))

	for _, pid := range allPids {
		// FIXME: if the process has halted and no process with this pid exists, we will get an error.
		err = time.ModifyTime(int(pid), int64(0), int64(0), 0)
		if err != nil {
			log.Error("failed to recover time", zap.Uint32("pid", pid), zap.Error(err))
			return errors.WithStack(err)
		}
	}

	return nil
}
