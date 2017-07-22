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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"

	"github.com/turbinelabs/api/client/tokencache"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/nonstdlib/flag/usage"
)

// TokenCachePath returns the path that the auth token should be cached at.
func TokenCachePath() string {
	return filepath.Join(os.Getenv("HOME"), ".tbnctl-auth-cache")
}

func populateDefaults(tc tokencache.TokenCache) tokencache.TokenCache {
	// kind of a question if we should make these values flagged
	tc.ClientID = "tbnctl"
	tc.ClientKey = "57f0c773-6fe8-4b56-9dd2-c915143a4c19"
	tc.ProviderURL = "https://login.turbinelabs.io/auth/realms/turbine-labs"
	return tc
}

type loginCfg struct {
	*globalConfigT

	user string
	pass string
}

type loginRunner struct {
	cfg *loginCfg
}

func login(tc tokencache.TokenCache, password string) error {
	cfg, err := tokencache.ToOAuthConfig(tc)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tkn, err := cfg.PasswordCredentialsToken(ctx, tc.Username, password)
	if err != nil {
		return fmt.Errorf("unable to authenticate using username %q and password: %v", tc.Username, err)
	}

	tc.SetToken(tokencache.WrapOAuth2Token(tkn))
	if err := tc.Save(TokenCachePath()); err != nil {
		return fmt.Errorf("unable to save new token: %v\n", err)
	}

	return nil
}

func readline() (string, error) {
	userBytes, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return "", err
	}

	lb := len(userBytes)
	str := string(userBytes)

	if lb == 0 {
		return "", nil
	}
	if str[lb-1] == '\n' {
		return str[0 : lb-1], nil
	}
	return str, nil
}

func (gc *loginRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	tokenCache, err := tokencache.NewFromFile(TokenCachePath())
	if err != nil {
		return cmd.Errorf(
			"Unable to process token cache (%v): %v",
			TokenCachePath(),
			err,
		)
	}

	if gc.cfg.user != "" {
		tokenCache.Username = gc.cfg.user
	} else {
		// username not set by flag, get interactively
		fmt.Printf("Username [%s]: ", tokenCache.Username)
		userBytes, err := readline()
		if err != nil {
			return cmd.Error(err)
		}
		if len(userBytes) != 0 {
			tokenCache.Username = strings.TrimSpace(string(userBytes))
		}
	}

	if len(tokenCache.Username) == 0 {
		return cmd.Errorf("Username must not be blank")
	}

	password := gc.cfg.pass
	if password == "" {
		// password not set by flag, get interactively
		fmt.Printf("Password: ")
		passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return cmd.Errorf("Unable to read password: %v\n", err)
		}
		password = string(passBytes)
		fmt.Println()
	}

	if len(password) == 0 {
		return cmd.Errorf("password must not be empty")
	}

	err = login(populateDefaults(tokenCache), password)
	if err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

func cmdLogin(cfg globalConfigT) *command.Cmd {
	runner := &loginRunner{&loginCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:    "login",
		Summary: "Obtain an authentication token that may be used for subsequent requests to the Turbine Labs API.",
		Usage:   "",
		Runner:  runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.user,
		"username",
		"",
		"the user name to login as",
	)

	cmd.Flags.StringVar(
		&runner.cfg.pass,
		"password",
		"",
		usage.Sensitive("password to be used when authenticating"),
	)

	return cmd
}
