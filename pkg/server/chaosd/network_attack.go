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

	"github.com/google/uuid"
	"github.com/pingcap/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/core"
	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
)

const (
	NetworkAttack = "network attack"
)

func (s *Server) NetworkAttack(attack *core.NetworkCommand) (string, error) {
	uid := uuid.New()
	ipsetName := ""
	if attack.NeedApplyIPSet() {
		ipset, err := attack.ToIPSet(fmt.Sprintf("chaos-%s", uid.String()[:16]))
		if err != nil {
			return "", errors.WithStack(err)
		}

		if err := flushIPSet(context.Background(), "", ipset); err != nil {
			return "", errors.WithStack(err)
		}
		ipsetName = ipset.Name
	}

	if err := s.exp.Set(context.Background(), &core.Experiment{
		Uid:            uid.String(),
		Status:         core.Created,
		Kind:           NetworkAttack,
		RecoverCommand: attack.String(),
	}); err != nil {
		return "", errors.WithStack(err)
	}

	switch attack.Action {
	case core.NetworkDelayAction:
		netem, err := attack.ToNetem()
		if err != nil {
			return "", errors.WithStack(err)
		}
		tc := &pb.Tc{
			Type:       pb.Tc_NETEM,
			Netem:      netem,
			Ipset:      ipsetName,
			Protocol:   attack.IPProtocol,
			SourcePort: attack.SourcePort,
			EgressPort: attack.EgressPort,
		}

		in := &pb.TcsRequest{
			Tcs: []*pb.Tc{tc},
		}

		if err := s.SetNodeTcRules(context.Background(), in); err != nil {
			return "", errors.WithStack(err)
		}
	}

	if err := s.exp.Update(context.Background(), uid.String(), core.Success, "", attack.String()); err != nil {
		return "", errors.WithStack(err)
	}

	return uid.String(), nil
}

func (s *Server) RecoverNetworkAttack(uid string, attack *core.NetworkCommand) error {
	switch attack.Action {
	case core.NetworkDelayAction:
		if err := s.SetNodeTcRules(context.Background(), &pb.TcsRequest{}); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := s.exp.Update(context.Background(), uid, core.Destroyed, "", attack.String()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Server) applyIPSet(attack *core.NetworkCommand) error {
	return nil
}
