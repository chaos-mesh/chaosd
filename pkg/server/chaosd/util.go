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
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/pingcap/log"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-daemon/pkg/bpm"
	"github.com/chaos-mesh/chaos-daemon/pkg/utils"
)

const (
	defaultProcPrefix = "/proc"
)

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ bpm.NsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", defaultProcPrefix, pid, string(typ))
}

// ReadCommName returns the command name of process
func ReadCommName(pid int) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/%d/comm", defaultProcPrefix, pid))
	if err != nil {
		return "", errors.WithStack(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(b), nil
}

// GetChildProcesses will return all child processes's pid. Include all generations.
// only return error when /proc/pid/tasks cannot be read
func GetChildProcesses(ppid uint32) ([]uint32, error) {
	procs, err := ioutil.ReadDir(defaultProcPrefix)
	if err != nil {
		return nil, errors.WithStack(err)
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
					log.Error("read status file error", zap.String("path", statusPath))
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
	return fmt.Errorf("error code: %v, msg: %s", err, string(output))
}
