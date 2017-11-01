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
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	apierror "github.com/turbinelabs/api/http/error"
	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/codec"
	"github.com/turbinelabs/nonstdlib/editor"
	"github.com/turbinelabs/nonstdlib/log/console"
	tbnos "github.com/turbinelabs/nonstdlib/os"
)

var objTypeList = []objecttype.ObjectType{
	objecttype.User,
	objecttype.Zone,
	objecttype.Proxy,
	objecttype.Domain,
	objecttype.Route,
	objecttype.SharedRules,
	objecttype.Cluster,
}

type Keyed interface {
	Key() string
	UpdateKey(string)
}

func updateKeyed(cmd *command.Cmd, src *[]string, tgt Keyed) command.CmdErr {
	if len(*src) > 0 {
		key, err := objKeyFromStrings(src)
		if err != nil {
			return cmd.BadInput(err)
		}

		if tgt.Key() != "" {
			console.Error().Printf("overwriting --key flag with object key from argument\n")
		}

		tgt.UpdateKey(key)
	}

	if tgt.Key() == "" {
		return cmd.BadInput(fmt.Errorf("expected object key, none found"))
	}

	return command.NoError()
}

func objKeyFromStrings(args *[]string) (string, error) {
	if len(*args) < 1 {
		return "", errors.New("expected object key missing")
	}

	key := (*args)[0]
	*args = (*args)[1:]

	return key, nil
}

func otFromStrings(args *[]string) (objecttype.ObjectType, error) {
	if len(*args) < 1 {
		return objecttype.ObjectType{}, errors.New("expected object type as first argument, got nothing")
	}

	ot, err := objecttype.FromName((*args)[0])
	if err != nil {
		return objecttype.ObjectType{}, fmt.Errorf("%s was not a valid object type", (*args)[0])
	}

	*args = (*args)[1:]

	return ot, nil
}

func objTypeNames() string {
	desc := []string{}
	for _, ot := range objTypeList {
		desc = append(desc, ot.Name)
	}

	return strings.Join(desc, ", ")
}

func createEditorHelp() string {
	return editorHelp("created")
}

func editingEditorHelp() string {
	return editorHelp("edited")
}

func editorHelp(act string) string {
	edtCmd, _, _ := editor.Get()
	return fmt.Sprintf(`{{bold "Editor Selection"}}

When changes need to be made an initial version of the object can be presented
in an editor. The command used to launch the editor is taken from the %s
environment variable and must block execution until the changes are saved and
the editor is closed. The current editor command is '%s'.

{{ul "Example EDITOR values"}}:

		vim

		emacs

		atom -w

{{bold "Using STDIN"}}

For scripting purposes it may be useful to use STDIN to provide the %s object
instead of using an interactive editor. If so, simply make the new version
available on STDIN through standard use of pipes.

{{ul "Example"}}: cat "new_cluster.json" | tbnctl create cluster`,
		editor.EditorVar,
		edtCmd,
		act,
	)
}

// UntypedSvc examines command args and gets an untyped interface supporting
// CRUD operations on one of the core objects underlying the Turbine Labs API.
//
// If an untyped service is not available an error will be returned.
func (gc *globalConfigT) UntypedSvc(args *[]string) (typelessIface, error) {
	ot, err := otFromStrings(args)
	if err != nil {
		return nil, err
	}

	svc := newTypelessIface(gc.apiClient, ot)
	if svc == nil {
		return nil, fmt.Errorf("Unsupported object type: %v\n", ot.Name)
	}

	return svc, nil
}

// PrintResult encodes an object per the configured codec and prints to stdout.
func (gc *globalConfigT) PrintResult(obj interface{}) {
	if err := gc.codec.Encode(obj, os.Stdout); err != nil {
		// make a best effort, use the string representation otherwise
		fmt.Println(obj)
	}
	// newline after object
	fmt.Println()
}

// editOrStdin is a helper function for commands that need to allow the user
// to modify some text starting from an encoded object or use stdin. If stdin
// has content then no object will be rendered and presented for modification.
// Otherwise fallback will be called and the returned object will be rendered
// using the codec available through gc and opened in an editor.
func editOrStdin(
	fallback func() (interface{}, error),
	gc *globalConfigT,
) (string, error) {
	txt, err := tbnos.ReadIfNonEmpty(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("could not process STDIN: %s", err.Error())
	}
	if txt != "" {
		return txt, nil
	}

	obj, err := fallback()
	if err != nil {
		return "", err
	}

	objstr, err := codec.EncodeToString(gc.codec, obj)
	if err != nil {
		return "", err
	}

	txt, err = editor.EditTextType(objstr, gc.codecFlags.Type())
	if err != nil {
		return "", err
	}

	return txt, nil
}

// PrettyCmdErr will marshal API client errors, and return the result of Error()
// for other errors.
func (gc *globalConfigT) PrettyCmdErr(cmd *command.Cmd, err error) command.CmdErr {
	switch e := err.(type) {
	case *apierror.Error:
		buf := &bytes.Buffer{}
		if err := gc.codec.Encode(e, buf); err != nil {
			return cmd.Error(err)
		}
		return cmd.Error(strings.TrimSpace(buf.String()))
	default:
		return cmd.Error(err)
	}
}
