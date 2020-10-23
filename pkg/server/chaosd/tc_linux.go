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

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/mock"
)

func applyTc(ctx context.Context, pid uint32, args ...string) error {
	// Mock point to return error in unit test
	if err := mock.On("TcApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return errors.WithStack(e)
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	nsPath := GetNsPath(pid, bpm.NetNS)

	cmd := bpm.DefaultProcessBuilder("tc", args...).SetNetNS(nsPath).SetContext(ctx).Build()
	log.Info("tc command", zap.String("command", cmd.String()), zap.Strings("args", args))

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Error("tc command error", zap.String("command", cmd.String()), zap.String("output", string(out)))
		return errors.WithStack(err)
	}

	return nil
}
