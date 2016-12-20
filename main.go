package main

import (
	goflag "flag"

	apiflag "github.com/turbinelabs/api/client/flags"
	"github.com/turbinelabs/cli"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/codec"
)

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

	gflags := &goflag.FlagSet{}
	globalConfig.apiFlags = apiflag.NewClientFromFlags(gflags)
	globalConfig.codecFlags = codec.NewFromFlags(gflags)

	app := cli.NewWithSubCmds(
		"Command line tool for interacting with the Turbine Labs API",
		"0.1",
		cmdGet(globalConfig),
		cmdEdit(globalConfig),
		cmdDelete(globalConfig),
		cmdList(globalConfig),
		cmdCreate(globalConfig),
		cmdProxyConfig(globalConfig),
	)
	app.SetFlags(gflags)

	app.Main()
}
