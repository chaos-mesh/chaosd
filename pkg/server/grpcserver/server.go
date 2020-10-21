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
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pingcap/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"

	"google.golang.org/grpc"

	grpcm "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/chaos-mesh/chaos-daemon/pkg/config"
	"github.com/chaos-mesh/chaos-daemon/pkg/server/chaosd"
	"github.com/chaos-mesh/chaos-daemon/pkg/server/pb"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
)

type grpcServer struct {
	conf  config.Config
	chaos *chaosd.Server
}

func NewServer(conf config.Config, chaos *chaosd.Server) *grpcServer {
	return &grpcServer{
		conf:  conf,
		chaos: chaos,
	}
}

func Register(s *grpcServer) {
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
	return nil, nil
}

func (s *grpcServer) FlushIPSets(ctx context.Context, in *pb.IPSetsRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *grpcServer) SetIptablesChains(ctx context.Context, in *pb.IptablesChainsRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *grpcServer) SetTimeOffset(ctx context.Context, in *pb.TimeRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *grpcServer) RecoverTimeOffset(ctx context.Context, in *pb.TimeRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *grpcServer) ContainerKill(ctx context.Context, in *pb.ContainerRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *grpcServer) ContainerGetPid(ctx context.Context, in *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	return nil, nil
}

func (s *grpcServer) ExecStressors(ctx context.Context, in *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	return nil, nil
}

func (s *grpcServer) CancelStressors(ctx context.Context, in *pb.CancelStressRequest) (*empty.Empty, error) {
	return nil, nil
}
