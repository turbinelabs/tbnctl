package main

import (
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
		gc.cfg.key == "",
		func() (interface{}, error) { return svc.Get(gc.cfg.key) },
		gc.cfg.globalConfigT,
	)

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
		Description: "object type is one of: " + objTypeNames(),
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
