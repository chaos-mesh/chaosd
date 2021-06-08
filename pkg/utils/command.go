package utils

import (
	"fmt"
	"os/exec"
	"reflect"
)

type Command struct {
	Name string
}

func (c Command) Unmarshal(val interface{}) *exec.Cmd {
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
	return exec.Command(c.Name, options...)
}
