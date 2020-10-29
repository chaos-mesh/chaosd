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

package config

import (
	"fmt"

	"github.com/pingcap/errors"
	flag "github.com/spf13/pflag"
)

// NewConfig creates a new config.
func NewConfig() *Config {
	cfg := &Config{}
	cfg.flagSet = flag.NewFlagSet("chaosd", flag.ContinueOnError)

	fg := cfg.flagSet

	fg.BoolVarP(&cfg.Version, "version", "V", false, "print version information and exit")
	fg.IntVarP(&cfg.ListenPort, "port", "p", 31767, "listen port of the Chaosd Server")
	fg.StringVarP(&cfg.ListenHost, "host", "h", "0.0.0.0", "listen host of the Chaosd Server")
	fg.StringVarP(&cfg.Runtime, "runtime", "r", "docker", "current container runtime")
	fg.BoolVar(&cfg.EnablePprof, "enable-pprof", true, "enable pprof")
	fg.IntVar(&cfg.PprofPort, "pprof-port", 31766, "listen port of the pprof server")
	fg.StringVarP(&cfg.Platform, "platform", "f", "local", "platform to deploy, default: local, supported platform: local, kubernetes")

	return cfg
}

// Config defines the configuration for Chaosd.
type Config struct {
	flagSet *flag.FlagSet

	Version bool

	ListenPort  int
	ListenHost  string
	Runtime     string
	EnablePprof bool
	PprofPort   int
	Platform    string
}

// Parse parses flag definitions from the argument list.
func (c *Config) Parse(arguments []string) error {
	if err := c.flagSet.Parse(arguments); err != nil {
		return errors.WithStack(err)
	}

	return c.Validate()
}

// Get the grpc address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.ListenHost, c.ListenPort)
}

// Validate is used to validate if some configurations are right.
func (c *Config) Validate() error {
	if !checkPlatform(c.Platform) {
		return errors.Errorf("platform %s is not supported", c.Platform)
	}

	if !checkRuntime(c.Runtime) {
		return errors.Errorf("container runtime %s is not supported", c.Runtime)
	}

	return nil
}

type Platform string

const (
	LocalPlatform      = "local"
	KubernetesPlatform = "kubernetes"
)

var supportPlatforms = []Platform{LocalPlatform, KubernetesPlatform}

// checkPlatform verifies if the platform is supported.
func checkPlatform(platform string) bool {
	for _, p := range supportPlatforms {
		if string(p) == platform {
			return true
		}
	}

	return false
}

var supportRuntimes = []string{"docker", "runtime"}

func checkRuntime(runtime string) bool {
	for _, r := range supportRuntimes {
		if r == runtime {
			return true
		}
	}

	return false
}
