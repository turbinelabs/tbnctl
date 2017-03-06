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
	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
)

type delCfg struct {
	*globalConfigT

	key string
}

type delRunner struct {
	cfg *delCfg
}

func (gc *delRunner) run(cmd *command.Cmd, args []string) error {
	svc, err := gc.cfg.UntypedSvc(args)
	if err != nil {
		return err
	}

	obj, err := svc.Get(gc.cfg.key)
	if err != nil {
		return err
	}

	e := svc.Delete(gc.cfg.key, svc.Checksum(obj))
	return gc.cfg.MkResult(obj, e)
}

func (gc *delRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	err := gc.run(cmd, args)
	if err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

func cmdDelete(cfg globalConfigT) *command.Cmd {
	runner := &delRunner{&delCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "delete",
		Summary:     "delete an object from Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.key,
		"key",
		"",
		tbnflag.Required("key of the object to delete"),
	)

	return cmd
}
