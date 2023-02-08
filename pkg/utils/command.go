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
	"reflect"
)

type Command struct {
	Name string
}

func (c Command) Unmarshal(val interface{}) *exec.Cmd {
	options := c.getOption(val)
	return exec.Command(c.Name, options...)
}

func (c Command) UnmarshalWithCtx(ctx context.Context, val interface{}) *exec.Cmd {
	options := c.getOption(val)
	return exec.CommandContext(ctx, c.Name, options...)
}

func (c Command) getOption(val interface{}) []string {
	v := reflect.ValueOf(val)

	var options []string
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get(c.Name)
		if v.Field(i).String() == "" || tag == "" {
			continue
		}

		if tag == "-" {
			options = append(options, v.Field(i).String())
		} else {
			options = append(options, fmt.Sprintf("%s=%v", tag, v.Field(i).String()))
		}
	}
	return options
}
