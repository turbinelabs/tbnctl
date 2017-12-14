package main

import (
	"os"

	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/nonstdlib/editor"
	tbnos "github.com/turbinelabs/nonstdlib/os"
)

const importZoneDesc = `Import a Zone to the Turbine Labs API. Typically the
input will be the output of a previous call to export-zone, with object keys
replaced by names. Referential integrity, assuming it is present in the input,
is maintained in the import. The Zone to be imported is assumed not to exist,
and import-zone will fail if the Zone is already present.`

func cmdImportZone(globalConfig globalConfigT) *command.Cmd {
	cmd := &command.Cmd{
		Name:        "import-zone",
		Summary:     "import a Zone to the Turbine Labs API",
		Usage:       "[OPTIONS] <zone-name>",
		Description: importZoneDesc,
	}

	cmd.Runner = &importZoneRunner{globalConfig}
	return cmd
}

type importZoneRunner struct {
	cfg globalConfigT
}

func (r *importZoneRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := r.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	return r.run(cmd, args)
}

func (r *importZoneRunner) run(cmd *command.Cmd, args []string) command.CmdErr {
	if len(args) != 1 {
		return cmd.BadInput("requires exactly one argument")
	}

	var (
		txt string
		err error
	)

	txt, err = tbnos.ReadIfNonEmpty(os.Stdin)
	if err != nil {
		return cmd.Errorf("could not process STDIN: %s", err)
	}
	if txt == "" {
		txt, err = editor.EditTextType("", r.cfg.codecFlags.Type())
		if err != nil {
			return cmd.Error(err)
		}
	}

	zo, err := importZone(r.cfg.apiClient, args[0], r.cfg.codec, txt)
	if err != nil {
		return r.cfg.PrettyCmdErr(cmd, err)
	}

	r.cfg.PrintResult(zo)

	return command.NoError()
}
