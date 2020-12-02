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
	"strings"

	"go.uber.org/zap"

	"github.com/pingcap/errors"
	"github.com/pingcap/log"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-daemon/pkg/server/serverpb"
)

const (
	ipsetExistErr        = "set with the same name already exists"
	ipExistErr           = "it's already added"
	ipsetNewNameExistErr = "a set with the new name already exists"
)

func (s *Server) FlushContainerIPSets(ctx context.Context, req *pb.IPSetsRequest) error {
	pid, err := s.criCli.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error("failed to get pid", zap.Error(err))
		return errors.WithStack(err)
	}

	nsPath := GetNsPath(pid, bpm.NetNS)

	for _, ipset := range req.Ipsets {
		err := flushIPSet(ctx, nsPath, ipset)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func flushIPSet(ctx context.Context, nsPath string, set *pb.IPSet) error {
	name := set.Name

	// If the ipset already exists, the ipset will be renamed to this temp name.
	tmpName := fmt.Sprintf("%sold", name)

	// the ipset while existing iptables rules are using them can not be deleted,.
	// so we creates an temp ipset and swap it with existing one.
	if err := createIPSet(ctx, nsPath, tmpName); err != nil {
		return errors.WithStack(err)
	}

	// add ips to the temp ipset
	if err := addCIDRsToIPSet(ctx, nsPath, tmpName, set.Cidrs); err != nil {
		return errors.WithStack(err)
	}

	// rename the temp ipset with the target name of ipset if the taget ipset not exists,
	// otherwise swap  them with each other.
	return renameIPSet(ctx, nsPath, tmpName, name)
}

func createIPSet(ctx context.Context, nsPath string, name string) error {
	// ipset name cannot be longer than 31 bytes
	if len(name) > 31 {
		name = name[:31]
	}

	cmd := bpm.DefaultProcessBuilder("ipset", "create", name, "hash:net").
		SetNetNS(nsPath).
		SetContext(ctx).
		Build()

	log.Debug("create ipset", zap.String("command", cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetExistErr) {
			log.Error("ipset create error",
				zap.String("command", cmd.String()), zap.String("output", output), zap.Error(err))
			return err
		}

		cmd := bpm.DefaultProcessBuilder("ipset", "flush", name).
			SetNetNS(nsPath).
			SetContext(ctx).
			Build()

		log.Debug("flush ipset", zap.String("command", cmd.String()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("ipset flush error",
				zap.String("command", cmd.String()), zap.String("output", string(out)), zap.Error(err))
			return err
		}
	}

	return nil
}

func addCIDRsToIPSet(ctx context.Context, nsPath string, name string, cidrs []string) error {
	for _, cidr := range cidrs {
		cmd := bpm.DefaultProcessBuilder("ipset", "add", name, cidr).SetNetNS(nsPath).SetContext(ctx).Build()

		log.Debug("add CIDR to ipset", zap.String("command", cmd.String()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, ipExistErr) {
				log.Error("ipset add error",
					zap.String("command", cmd.String()), zap.String("output", output), zap.Error(err))
				return err
			}
		}
	}

	return nil
}

func renameIPSet(ctx context.Context, nsPath string, oldName string, newName string) error {
	cmd := bpm.DefaultProcessBuilder("ipset", "rename", oldName, newName).SetNetNS(nsPath).SetContext(ctx).Build()

	log.Debug("rename ipset", zap.String("command", cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetNewNameExistErr) {
			log.Error("rename ipset failed",
				zap.String("command", cmd.String()), zap.String("output", output), zap.Error(err))
			return errors.WithStack(err)
		}

		// swap the old ipset and the new ipset if the new ipset already exist.
		cmd := bpm.DefaultProcessBuilder("ipset", "swap", oldName, newName).SetNetNS(nsPath).SetContext(ctx).Build()

		log.Debug("swap ipset", zap.String("command", cmd.String()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("swap ipset failed",
				zap.String("command", cmd.String()), zap.String("output", string(out)), zap.Error(err))
			return errors.WithStack(err)
		}
	}
	return nil
}
