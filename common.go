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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/codec"
	"github.com/turbinelabs/nonstdlib/editor"
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

func otFromStrings(args []string) (objecttype.ObjectType, error) {
	if len(args) < 1 {
		return objecttype.ObjectType{}, errors.New("expected object type as first argument, got nothing")
	}
	ot, err := objecttype.FromName(args[0])
	if err != nil {
		return objecttype.ObjectType{}, fmt.Errorf("%s was not a valid object type", args[0])
	}

	return ot, nil
}

func objTypeNames() string {
	desc := []string{}
	for _, ot := range objTypeList {
		desc = append(desc, ot.Name)
	}

	return strings.Join(desc, ", ")
}

// UntypedSvc examines command args and gets an untyped interface supporting
// CRUD operations on one of the core objects underlying the Turbine Labs API.
//
// If an untyped service is not available an error will be returned.
func (gc *globalConfigT) UntypedSvc(args []string) (typelessIface, error) {
	ot, err := otFromStrings(args)
	if err != nil {
		return nil, err
	}

	svc := newTypelessIface(gc.apiClient, ot)
	if svc == nil {
		return nil, fmt.Errorf("Unsupported object type: %v\n", args[0])
	}

	return svc, nil
}

// MkResult takes an object resulting from the operation and any error that
// was encountered. The error (if not nil) or object (if err is nil) will be
// encoded per the configured codec and printed to stdout. An error will be
// returned only if there was a problem encoding obj or err.
func (gc *globalConfigT) MkResult(obj interface{}, err error) error {
	if err != nil {
		if eerr := gc.codec.Encode(err, os.Stdout); eerr != nil {
			return eerr
		}
		return nil
	}

	if err := gc.codec.Encode(obj, os.Stdout); err != nil {
		return err
	}

	return nil
}

// editOrStdin is a helper function for commands that need to allow the user
// to modify some text starting from an encoded object or use stdin. If stdin
// is set then no object will be rendered and presented for modification.
// Otherwise fallback will be called and the returned object will be rendered
// using the codec available through gc and opened in an editor.
func editOrStdin(
	stdin bool,
	fallback func() (interface{}, error),
	gc *globalConfigT,
) (string, error) {
	txt := ""

	if stdin {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		txt = string(b)
	} else {
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
	}

	return txt, nil
}
