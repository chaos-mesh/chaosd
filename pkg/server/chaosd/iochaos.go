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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pingcap/log"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/pb"
)

const (
	todaBin = "/usr/local/bin/toda"
)

func (s *Server) ApplyIoChaos(ctx context.Context, in *pb.ApplyIoChaosRequest) (*pb.ApplyIoChaosResponse, error) {
	if in.Instance != 0 {
		err := s.killIoChaos(ctx, in.Instance, in.StartTime)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	actions := []v1alpha1.IoChaosAction{}
	json.Unmarshal([]byte(in.Actions), &actions)
	log.Info("the length of actions", zap.Int("length", len(actions)))
	if len(actions) == 0 {
		return &pb.ApplyIoChaosResponse{
			Instance:  0,
			StartTime: 0,
		}, nil
	}

	pid, err := s.criCli.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		log.Error("failed to get pid", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	// TODO: make this log level configurable
	args := fmt.Sprintf("--path %s --pid %d --verbose info", in.Volume, pid)
	log.Info("executing", zap.String("cmd", todaBin+" "+args))
	cmd := bpm.DefaultProcessBuilder(todaBin, strings.Split(args, " ")...).
		EnableSuicide().
		SetIdentifier(in.ContainerId).
		Build()
	cmd.Stdin = strings.NewReader(in.Actions)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ct, err := procState.CreateTime()
	if err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error("kill toda failed", zap.Any("request", in), zap.Error(kerr))
		}
		return nil, errors.WithStack(err)
	}

	return &pb.ApplyIoChaosResponse{
		Instance:  int64(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *Server) killIoChaos(ctx context.Context, pid int64, startTime int64) error {
	log.Info("killing toda", zap.Int64("pid", pid))

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(pid), startTime)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Info("kill toda successfully")
	return nil
}
