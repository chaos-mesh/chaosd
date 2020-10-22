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

	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/pb"
)

func (s *Server) ExecContainerStress(context.Context, *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	return nil, nil
}

func (s *Server) CancelContainerStress(context.Context, *pb.CancelStressRequest) error {
	return nil
}
