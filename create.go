package main

import (
	"github.com/turbinelabs/cli/command"
)

type createCfg struct {
	*globalConfigT

	useStdin bool
}

type createRunner struct {
	cfg *createCfg
}

func (gc *createRunner) run(svc typelessIface) error {
	txt, err := editOrStdin(
		gc.cfg.useStdin,
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
	return gc.cfg.MkResult(svc.Create(dest))
}

func (gc *createRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
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

func cmdCreate(cfg globalConfigT) *command.Cmd {
	runner := &createRunner{&createCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "create",
		Summary:     "create an object within Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.BoolVar(
		&runner.cfg.useStdin,
		"stdin",
		false,
		`If this flag is set definition for the new object will be taken from
stdin instead of opening an editor.`,
	)

	return cmd
}
