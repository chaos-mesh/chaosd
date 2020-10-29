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

package grpcserver

import (
	"context"
	"net"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/golang/protobuf/ptypes/empty"
	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/server/chaosd"
	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
)

type grpcServer struct {
	conf  *config.Config
	chaos *chaosd.Server
}

func NewServer(conf *config.Config, chaos *chaosd.Server) *grpcServer {
	return &grpcServer{
		conf:  conf,
		chaos: chaos,
	}
}

func Register(s *grpcServer) {
	if s.conf.Platform != config.KubernetesPlatform {
		return
	}

	opts := []grpc.ServerOption{
		grpcm.WithUnaryServerChain(
			utils.TimeoutServerInterceptor,
		),
	}

	gs := grpc.NewServer(opts...)

	pb.RegisterChaosDaemonServer(gs, s)
	reflection.Register(gs)

	go func() {
		addr := s.conf.Address()
		log.Info("starting GRPC endpoint", zap.String("address", addr))
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal("failed to listen GRPC address", zap.Error(err))
		}

		if err := gs.Serve(listener); err != nil {
			gs.Stop()
			log.Fatal("failed to start GRPC endpoint", zap.Error(err))
		}
	}()
}

func (s *grpcServer) SetTcs(ctx context.Context, in *pb.TcsRequest) (*empty.Empty, error) {
	log.Info("handle tc request", zap.Any("request", in))
	if err := s.chaos.SetContainerTcRules(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) FlushIPSets(ctx context.Context, in *pb.IPSetsRequest) (*empty.Empty, error) {
	log.Info("flush ipset", zap.Any("request", in))
	if err := s.chaos.FlushContainerIPSets(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) SetIptablesChains(ctx context.Context, in *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log.Info("set iptables chains", zap.Any("request", in))
	if err := s.chaos.SetContainerIptablesChains(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) SetTimeOffset(ctx context.Context, in *pb.TimeRequest) (*empty.Empty, error) {
	log.Info("shift time", zap.Any("request", in))
	if err := s.chaos.SetContainerTime(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) RecoverTimeOffset(ctx context.Context, in *pb.TimeRequest) (*empty.Empty, error) {
	log.Info("recover time", zap.Any("request", in))
	if err := s.chaos.RecoverContainerTime(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) ContainerKill(ctx context.Context, in *pb.ContainerRequest) (*empty.Empty, error) {
	log.Info("kill container", zap.Any("request", in))

	if err := s.chaos.ContainerKill(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) ContainerGetPid(ctx context.Context, in *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	log.Info("get container pid", zap.Any("request", in))

	return s.chaos.ContainerGetPid(ctx, in)
}

func (s *grpcServer) ExecStressors(ctx context.Context, in *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log.Info("execute stress", zap.Any("request", in))

	return s.chaos.ExecContainerStress(ctx, in)
}

func (s *grpcServer) CancelStressors(ctx context.Context, in *pb.CancelStressRequest) (*empty.Empty, error) {
	log.Info("cancel stress", zap.Any("request", in))

	if err := s.chaos.CancelContainerStress(ctx, in); err != nil {
		return nil, errors.WithStack(err)
	}
	return &empty.Empty{}, nil
}

func (s *grpcServer) ApplyIoChaos(ctx context.Context, in *pb.ApplyIoChaosRequest) (*pb.ApplyIoChaosResponse, error) {
	log.Info("apply iochaos", zap.Any("request", in))
	return s.chaos.ApplyIoChaos(ctx, in)
}
