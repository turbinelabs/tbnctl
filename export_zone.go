package main

import (
	"github.com/turbinelabs/cli/command"
)

const exportZoneDesc = `Export a Zone from the Turbine Labs API. Object keys are
replaced with human-readable names, and referential integrity is maintained. The
output of export-zone is suitable for input into import-zone.`

func cmdExportZone(globalConfig globalConfigT) *command.Cmd {
	cmd := &command.Cmd{
		Name:        "export-zone",
		Summary:     "export a Zone from the Turbine Labs API",
		Usage:       "[OPTIONS] <zone-name>|<zone-key>",
		Description: exportZoneDesc,
	}

	cmd.Runner = &exportZoneRunner{globalConfig}
	return cmd
}

type exportZoneRunner struct {
	cfg globalConfigT
}

func (r *exportZoneRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := r.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	return r.run(cmd, args)
}

func (r *exportZoneRunner) run(cmd *command.Cmd, args []string) command.CmdErr {
	if len(args) != 1 {
		return cmd.BadInput("requires exactly one argument")
	}

	zo, err := exportZone(r.cfg.apiClient, args[0])
	if err != nil {
		return r.cfg.PrettyCmdErr(cmd, err)
	}

	r.cfg.PrintResult(zo)

	return command.NoError()
}
