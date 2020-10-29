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
	"fmt"

	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
)

// ContainerKill kills container according to container id in the req
func (s *Server) ContainerKill(ctx context.Context, req *pb.ContainerRequest) error {
	action := req.Action.Action
	if action != pb.ContainerAction_KILL {
		err := fmt.Errorf("container action is %s , not kill", pb.ContainerAction_Action_name[int32(action)])
		log.Error("container action is not expected", zap.Error(err))
		return errors.WithStack(err)
	}

	err := s.criCli.ContainerKillByContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error("failed to kill container", zap.Error(err))
		return errors.WithStack(err)
	}

	return nil
}

func (s *Server) ContainerGetPid(ctx context.Context, req *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	action := req.Action.Action
	if action != pb.ContainerAction_GETPID {
		err := fmt.Errorf("container action is %s , not getpid", pb.ContainerAction_Action_name[int32(action)])
		log.Error("container action is not expected", zap.Error(err))
		return nil, err
	}

	pid, err := s.criCli.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error("failed to get container pid", zap.Error(err))
		return nil, err
	}

	return &pb.ContainerResponse{Pid: pid}, nil
}
