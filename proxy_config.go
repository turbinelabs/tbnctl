/*
Copyright 2017 Turbine Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/tbnproxy/confagent"
	"github.com/turbinelabs/tbnproxy/confagent/nginxconfig"
	"github.com/turbinelabs/tbnproxy/configwriter"
)

func cmdProxyConfig(cfg globalConfigT) *command.Cmd {
	runner := &pcRunner{&pcConfig{}, nil}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "proxy-config",
		Summary:     "get the proxy config",
		Usage:       "<proxy key> | <zone name> <proxy name>",
		Description: "print the config that is used by a proxy server",
		Runner:      runner,
	}

	runner.cfg.cwFlags = configwriter.NewFromFlags(
		&cmd.Flags, configwriter.LogDirMayNotExist())

	return cmd
}

type pcConfig struct {
	*globalConfigT
	cwFlags configwriter.FromFlags

	zoneName  string
	proxyName string
	proxyKey  api.ProxyKey
}

func (c *pcConfig) Validate(args []string) error {
	la := len(args)
	if la == 0 || la > 2 {
		return errors.New(
			"arguments must be either <proxy key> or <zone name> <proxy name>")
	}

	if la == 1 {
		c.proxyKey = api.ProxyKey(args[0])
	}

	if la == 2 {
		c.zoneName = args[0]
		c.proxyName = args[1]
	}

	if err := c.cwFlags.Validate(); err != nil {
		return err
	}

	return c.globalConfigT.Validate()
}

func (c *pcConfig) Make() (confagent.ConfAgent, error) {
	err := c.globalConfigT.Make()
	if err != nil {
		return nil, err
	}

	var lookupFn confagent.ProxyKeyLookup
	if c.proxyKey != "" {
		lookupFn = confagent.ProxyKeyLookupNoop(c.proxyKey)
	} else {
		lookupFn = confagent.ProxyKeyFromZoneProxyNames(c.zoneName, c.proxyName)
	}

	cw, err := c.cwFlags.Make()
	if err != nil {
		return nil, fmt.Errorf(
			"Unable to construct configwriter: %v",
			err.Error())
	}

	agent, err := confagent.New(
		c.apiClient.All,
		cw,
		nginxconfig.Default(),
		lookupFn,
		1*time.Hour,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to construct config agent: %v", err.Error())
	}

	return agent, nil
}

type pcRunner struct {
	cfg *pcConfig

	agent confagent.Poller
}

func (r *pcRunner) Prepare(
	cmd *command.Cmd,
	args []string,
) (confagent.Poller, command.CmdErr) {
	if err := r.cfg.Validate(args); err != nil {
		return nil, cmd.BadInput(err)
	}

	agent, err := r.cfg.Make()
	if err != nil {
		return nil, cmd.Error(err)
	}

	return agent, command.NoError()
}

func (r pcRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	agent, err := r.Prepare(cmd, args)
	if err != command.NoError() {
		return err
	}

	if err := agent.Poll(); err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}
