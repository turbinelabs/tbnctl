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
	"fmt"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/cli/command"
)

type keyCfg struct {
	*globalConfigT
}

type keyRunner struct {
	cfg *keyCfg
}

func (kr *keyRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := kr.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	if len(args) < 1 {
		return cmd.BadInput("must specify an access-tokens command")
	}

	svc := kr.cfg.apiClient.Admin.AccessToken()

	var (
		res interface{}
		err error
	)

	switch args[0] {
	case "list":
		res, err = svc.Index()

	case "add":
		if len(args) < 2 {
			return cmd.BadInput("description should be provided summarizing intended token use")
		}
		res, err = svc.Create(api.AccessToken{Description: args[1]})

	case "remove":
		if len(args) < 2 {
			return cmd.BadInput("access token to be removed must be specified")
		}

		key := api.AccessTokenKey(args[1])
		curs, err := svc.Index(service.AccessTokenFilter{AccessTokenKey: key})
		if err != nil {
			return kr.cfg.PrettyCmdErr(cmd, err)
		} else if len(curs) != 1 {
			return cmd.Errorf("unable to locate access token %v", args[1])
		}

		cur := curs[0]
		err = svc.Delete(key, cur.Checksum)

	default:
		return cmd.BadInput(fmt.Sprintf("%q is not a valid access-tokens command", args[0]))
	}

	if err != nil {
		return kr.cfg.PrettyCmdErr(cmd, err)
	}
	kr.cfg.PrintResult(res)

	return command.NoError()
}

func cmdTokens(cfg globalConfigT) *command.Cmd {
	runner := &keyRunner{&keyCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:    "access-tokens",
		Summary: "manage AccessTokens for your account",
		Usage:   "<command>",
		Description: `Provides management commands for access tokens associated with this account.

Commands available are:

  {{ul "list"}}:
    list all access tokens associated with this account

  {{ul "add <comment>"}}:
    adds a new access tokens to this account with the specified comment

  {{ul "remove <key>"}}:
    removes the specified key from this account
`,
		Runner: runner,
	}

	return cmd
}
