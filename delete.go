package main

import (
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/nonstdlib/flag"
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
		flag.Required("key of the object to delete"),
	)

	return cmd
}
