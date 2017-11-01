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
)

type createCfg struct {
	*globalConfigT
}

type createRunner struct {
	cfg *createCfg
}

func (gc *createRunner) run(svc typelessIface) error {
	txt, err := editOrStdin(
		func() (interface{}, error) { return svc.Zero(), nil },
		gc.cfg.globalConfigT,
	)
	if err != nil {
		return err
	}

	dest, err := svc.ObjFromString(txt, gc.cfg.codec)
	if err != nil {
		return err
	}
	obj, err := svc.Create(dest)
	if err != nil {
		return err
	}
	gc.cfg.PrintResult(obj)
	return nil
}

func (gc *createRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	svc, err := gc.cfg.UntypedSvc(&args)
	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}

	err = gc.run(svc)
	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}

	return command.NoError()
}

func cmdCreate(cfg globalConfigT) *command.Cmd {
	runner := &createRunner{&createCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "create",
		Summary:     "create an object within Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames() + "\n\n" + createEditorHelp(),
		Runner:      runner,
	}

	return cmd
}
