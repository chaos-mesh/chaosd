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

package utils

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// CommandPools is a group of commands runner
type CommandPools struct {
	cancel context.CancelFunc
	pools  *tunny.Pool
	wg     sync.WaitGroup
}

// NewCommandPools returns a new CommandPools
func NewCommandPools(ctx context.Context, deadline *time.Time, size int) *CommandPools {
	var ctx2 context.Context
	var cancel context.CancelFunc
	if deadline != nil {
		ctx2, cancel = context.WithDeadline(ctx, *deadline)
	} else {
		ctx2, cancel = context.WithCancel(ctx)
	}
	return &CommandPools{
		cancel: cancel,
		pools: tunny.NewFunc(size, func(payload interface{}) interface{} {
			cmdPayload, ok := payload.(lo.Tuple2[string, []string])
			if !ok {
				return mo.Err[[]byte](fmt.Errorf("payload is not CommandPayload"))
			}
			name, args := cmdPayload.Unpack()
			cmd := exec.CommandContext(ctx2, name, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return mo.Err[[]byte](fmt.Errorf("%s: %s", err, string(output)))
			}
			return mo.Ok[[]byte](output)
		}),
	}
}

func (p *CommandPools) Process(name string, args []string) ([]byte, error) {
	result, ok := p.pools.Process(lo.Tuple2[string, []string]{
		A: name,
		B: args,
	}).(mo.Result[[]byte])
	if !ok {
		return nil, fmt.Errorf("payload is not Result[[]byte]")
	}
	return result.Get()
}

func (p *CommandPools) Start(name string, args []string, outputHandler func([]byte, error)) {
	p.wg.Add(1)
	go func() {
		output, err := p.Process(name, args)
		outputHandler(output, err)
		p.wg.Done()
	}()
}

func (p *CommandPools) Wait() {
	p.wg.Wait()
}

func (p *CommandPools) Close() {
	p.cancel()
	p.Wait()
	p.pools.Close()
}
