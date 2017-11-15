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

type delCfg struct {
	*globalConfigT

	key  string
	deep bool
}

func (dc *delCfg) Key() string         { return dc.key }
func (dc *delCfg) UpdateKey(nk string) { dc.key = nk }

type delRunner struct {
	cfg *delCfg
}

func (gc *delRunner) run(cmd *command.Cmd, args []string) command.CmdErr {
	svc, err := gc.cfg.UntypedSvc(&args)
	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}

	if cerr := updateKeyed(cmd, &args, gc.cfg); cerr != command.NoError() {
		return cerr
	}

	obj, err := svc.Get(gc.cfg.key)
	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}

	if gc.cfg.deep {
		err = svc.DeepDelete(gc.cfg.key, svc.Checksum(obj), gc.cfg.apiClient)
	} else {
		err = svc.Delete(gc.cfg.key, svc.Checksum(obj))
	}

	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}
	gc.cfg.PrintResult(obj)

	return command.NoError()
}

func (gc *delRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	return gc.run(cmd, args)
}

func cmdDelete(cfg globalConfigT) *command.Cmd {
	runner := &delRunner{&delCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "delete",
		Summary:     "delete an object from Turbine Labs API",
		Usage:       "[OPTIONS] <object type> <object key>",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.key,
		"key",
		"",
		"[deprecated] key of the object to delete",
	)

	cmd.Flags.BoolVar(
		&runner.cfg.deep,
		"deep",
		false,
		"if true, delete the entire object graph below the specified object",
	)

	return cmd
}
