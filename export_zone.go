/*
Copyright 2018 Turbine Labs, Inc.

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
