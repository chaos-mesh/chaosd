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
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/pingcap/log"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/mock"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
)

const (
	pausePath         = "/usr/local/bin/pause"
	suicidePath       = "/usr/local/bin/suicide"
	defaultProcPrefix = "/proc"
)

type nsType string

const (
	mountNS nsType = "mnt"
	utsNS   nsType = "uts"
	ipcNS   nsType = "ipc"
	netNS   nsType = "net"
	pidNS   nsType = "pid"
	userNS  nsType = "user"
)

var nsArgMap = map[nsType]string{
	mountNS: "m",
	utsNS:   "u",
	ipcNS:   "i",
	netNS:   "n",
	pidNS:   "p",
	userNS:  "U",
}

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ nsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", defaultProcPrefix, pid, string(typ))
}

// processBuilder builds a exec.Cmd for daemon
type processBuilder struct {
	cmd  string
	args []string

	nsOptions []nsOption

	pause   bool
	suicide bool
}

func defaultProcessBuilder(cmd string, args ...string) *processBuilder {
	return &processBuilder{
		cmd:       cmd,
		args:      args,
		nsOptions: []nsOption{},
		pause:     false,
		suicide:   false,
	}
}

func (b *processBuilder) SetNetNS(nsPath string) *processBuilder {
	return b.SetNS([]nsOption{{
		Typ:  netNS,
		Path: nsPath,
	}})
}

func (b *processBuilder) SetPidNS(nsPath string) *processBuilder {
	return b.SetNS([]nsOption{{
		Typ:  pidNS,
		Path: nsPath,
	}})
}

func (b *processBuilder) SetNS(options []nsOption) *processBuilder {
	b.nsOptions = append(b.nsOptions, options...)

	return b
}

func (b *processBuilder) EnablePause() *processBuilder {
	b.pause = true

	return b
}

func (b *processBuilder) EnableSuicide() *processBuilder {
	b.suicide = true

	return b
}

func (b *processBuilder) Build(ctx context.Context) *exec.Cmd {
	// The call routine is pause -> suicide -> nsenter --(fork)-> suicide -> process
	// so that when chaos-daemon killed the suicide process, the sub suicide process will
	// receive a signal and exit.
	// For example:
	// If you call `nsenter -p/proc/.../ns/pid bash -c "while true; do sleep 1; date; done"`
	// then even you kill the nsenter process, the subprocess of it will continue running
	// until it gets killed. The suicide program is used to make sure that the subprocess will
	// be terminated when its parent died.
	// But the `./bin/suicide nsenter -p/proc/.../ns/pid ./bin/suicide bash -c "while true; do sleep 1; date; done"`
	// can fix this problem. The first suicide is used to ensure when chaos-daemon is dead, the process is killed

	// I'm not sure this method is 100% reliable, but half a loaf is better than none.

	args := b.args
	cmd := b.cmd

	if b.suicide {
		args = append([]string{cmd}, args...)
		cmd = suicidePath
	}

	if len(b.nsOptions) > 0 {
		args = append([]string{"--", cmd}, args...)
		for _, option := range b.nsOptions {
			args = append([]string{"-" + nsArgMap[option.Typ] + option.Path}, args...)
		}
		cmd = "nsenter"
	}

	if b.suicide {
		args = append([]string{cmd}, args...)
		cmd = suicidePath
	}

	if b.pause {
		args = append([]string{cmd}, args...)
		cmd = pausePath
	}

	if c := mock.On("MockProcessBuild"); c != nil {
		f := c.(func(context.Context, string, ...string) *exec.Cmd)
		return f(ctx, cmd, args...)
	}

	log.Info("build command", zap.String("command", cmd+" "+strings.Join(args, " ")))
	return exec.CommandContext(ctx, cmd, args...)
}

type nsOption struct {
	Typ  nsType
	Path string
}

// ReadCommName returns the command name of process
func ReadCommName(pid int) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/%d/comm", defaultProcPrefix, pid))
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// GetChildProcesses will return all child processes's pid. Include all generations.
// only return error when /proc/pid/tasks cannot be read
func GetChildProcesses(ppid uint32) ([]uint32, error) {
	procs, err := ioutil.ReadDir(defaultProcPrefix)
	if err != nil {
		return nil, err
	}

	type processPair struct {
		Pid  uint32
		Ppid uint32
	}

	pairs := make(chan processPair)
	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup

		for _, proc := range procs {
			_, err := strconv.ParseUint(proc.Name(), 10, 32)
			if err != nil {
				continue
			}

			statusPath := defaultProcPrefix + "/" + proc.Name() + "/stat"

			wg.Add(1)
			go func() {
				defer wg.Done()

				reader, err := os.Open(statusPath)
				if err != nil {
					log.Error("read status file error", zap.String("path", statusPath), zap.Error(err))
					return
				}

				var (
					pid    uint32
					comm   string
					state  string
					parent uint32
				)
				// according to procfs's man page
				fmt.Fscanf(reader, "%d %s %s %d", &pid, &comm, &state, &parent)

				pairs <- processPair{
					Pid:  pid,
					Ppid: parent,
				}
			}()
		}

		wg.Wait()
		done <- true
	}()

	processGraph := utils.NewGraph()
	for {
		select {
		case pair := <-pairs:
			processGraph.Insert(pair.Ppid, pair.Pid)
		case <-done:
			return processGraph.Flatten(ppid), nil
		}
	}
}

func encodeOutputToError(output []byte, err error) error {
	return errors.Errorf("error code: %v, msg: %s", err, string(output))
}
