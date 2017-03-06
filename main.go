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
	goflag "flag"

	apiflag "github.com/turbinelabs/api/client/flags"
	"github.com/turbinelabs/cli"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/codec"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
)

const TbnPublicVersion = "0.4.1"

var cmds = []func(globalConfigT) *command.Cmd{
	cmdList,
	cmdGet,
	cmdCreate,
	cmdEdit,
	cmdDelete,
	cmdInitZone,
}

type globalConfigT struct {
	apiFlags   apiflag.ClientFromFlags
	apiClient  *unifiedSvc
	codecFlags codec.FromFlags
	codec      codec.Codec
}

// Prepare handles getting everything validated, instantiated, and set on the
// globalConfigT. It returns an error if any configuration is invalid or fails
// to produce the expected component.
func (gc *globalConfigT) Prepare(cmd *command.Cmd) command.CmdErr {
	if err := gc.Validate(); err != nil {
		return cmd.BadInput(err)
	}
	if err := gc.Make(); err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

// Validate calls Validate on the nested flag-configured components and returns
// an error if any of them fail to validate.
func (gc globalConfigT) Validate() error {
	if err := gc.apiFlags.Validate(); err != nil {
		return err
	}

	if err := gc.codecFlags.Validate(); err != nil {
		return err
	}

	return nil
}

// Make calls Make on nested flag-configured components and saves the produced
// objects into the config object. It returns an error if any of the configured
// children return an error.
func (gc *globalConfigT) Make() error {
	svc, err := gc.apiFlags.Make()
	if err != nil {
		return err
	}
	svca, err := gc.apiFlags.MakeAdmin()
	if err != nil {
		return err
	}

	gc.apiClient = &unifiedSvc{svc, svca}
	gc.codec = gc.codecFlags.Make()

	return nil
}

func main() {
	globalConfig := globalConfigT{}

	gflags := tbnflag.Wrap(&goflag.FlagSet{})
	globalConfig.apiFlags = apiflag.NewClientFromFlags(gflags.Scope("api", "API"))
	globalConfig.codecFlags = codec.NewFromFlags(gflags)

	subs := []*command.Cmd{}
	for _, mkCmd := range cmds {
		subs = append(subs, mkCmd(globalConfig))
	}

	app := cli.NewWithSubCmds(
		"Command line tool for interacting with the Turbine Labs API",
		TbnPublicVersion,
		subs[0],
		subs[1],
		subs[2:]...,
	)
	app.SetFlags(gflags.Unwrap())

	app.Main()
}
