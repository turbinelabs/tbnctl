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

	"github.com/turbinelabs/cli/command"
)

type editCfg struct {
	*globalConfigT

	key string
}

type editRunner struct {
	cfg *editCfg
}

func (gc *editRunner) run(svc typelessIface) error {
	objstr, err := editOrStdin(
		func() (interface{}, error) {
			if gc.cfg.key == "" {
				return nil, errors.New("object key must be specified")
			}
			return svc.Get(gc.cfg.key)
		},
		gc.cfg.globalConfigT,
	)
	if err != nil {
		return err
	}

	dest, err := svc.ObjFromString(objstr, gc.cfg.codec)
	if err != nil {
		return err
	}

	return gc.cfg.MkResult(svc.Modify(dest))
}

func (gc *editRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	svc, err := gc.cfg.UntypedSvc(args)
	if err != nil {
		return cmd.Error(err)
	}

	err = gc.run(svc)
	if err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

func cmdEdit(cfg globalConfigT) *command.Cmd {
	runner := &editRunner{&editCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "edit",
		Summary:     "edit an object from Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames() + "\n\n" + editingEditorHelp(),
		Runner:      runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.key,
		"key",
		"",
		"key of the object to retrieve, if not provided will read input from stdin",
	)

	return cmd
}
