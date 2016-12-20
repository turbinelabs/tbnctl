package main

import (
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/nonstdlib/flag"
)

type getCfg struct {
	*globalConfigT

	key string
}

func (c getCfg) Validate() error {
	return c.globalConfigT.Validate()
}

func (c getCfg) Make() error {
	return c.globalConfigT.Make()
}

type getRunner struct {
	cfg *getCfg
}

func (gc *getRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	svc, err := gc.cfg.UntypedSvc(args)
	if err != nil {
		return command.Error(err)
	}

	err = gc.cfg.MkResult(svc.Get(gc.cfg.key))
	if err != nil {
		return command.Error(err)
	}
	return command.NoError()
}

func cmdGet(cfg globalConfigT) *command.Cmd {
	runner := &getRunner{&getCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "get",
		Summary:     "retrieve an object from Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.key,
		"key",
		"",
		flag.Required("key of the object to retrieve"),
	)

	return cmd
}
