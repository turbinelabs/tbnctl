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
	"github.com/turbinelabs/api/client/tokencache"
	"github.com/turbinelabs/cli/command"
)

type logoutRunner struct{}

func (gc *logoutRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	tokenCache, err := tokencache.NewFromFile(TokenCachePath())
	if err != nil {
		return cmd.Errorf(
			"Unable to process token cache (%v): %v",
			TokenCachePath(),
			err,
		)
	}

	tokenCache.SetToken(nil)
	if err := tokenCache.Save(TokenCachePath()); err != nil {
		return cmd.Errorf("Unable to invalidate cached auth token")
	}

	return command.NoError()
}

func cmdLogout(cfg globalConfigT) *command.Cmd {
	runner := &logoutRunner{}

	cmd := &command.Cmd{
		Name:    "logout",
		Summary: "Invalidate an authentication token proviously saved through the login command",
		Usage:   "",
		Runner:  runner,
	}

	return cmd
}
