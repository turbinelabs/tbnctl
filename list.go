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
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/cli/command"
	tbnstrings "github.com/turbinelabs/nonstdlib/strings"
	tbntabwriter "github.com/turbinelabs/nonstdlib/text/tabwriter"
)

type listCfg struct {
	*globalConfigT

	fmt              string
	fmtHeader        string
	showFilterFields bool
	sliceSep         string
}

type listRunner struct {
	cfg *listCfg
}

var zoneName = "{{with $z := getZone .ZoneKey}}{{$z.Name}}{{end}}"
var domainNamePort = "{{with $d := getDomain .DomainKey}}{{$d.Name}}:{{$d.Port}}{{end}}"

var headerStrings = map[string]map[string]string{
	objecttype.Cluster.Name: {
		"summary": "Cluster Key\tInstances\tZone\tName",
	},
	objecttype.SharedRules.Name: {
		"summary": "SharedRulesKey\tZoneKey\tName",
	},
	objecttype.Route.Name: {
		"summary":   "Route Key\tPath}}\tName:port\tZone",
		"path-only": "Route Key\tPath",
	},
	objecttype.User.Name: {
		"summary": "User Key\tEmail",
	},
}

var formatStrings = map[string]map[string]string{
	objecttype.Cluster.Name: {
		"summary": "{{.ClusterKey}}\t{{len .Instances}}\t" + zoneName + "\t{{.Name}}",
	},
	objecttype.SharedRules.Name: {
		"summary": "{{.SharedRulesKey}}\t" + zoneName + "\t{{.Name}}",
	},
	objecttype.Route.Name: {
		"summary":   "{{.RouteKey}}\t{{.Path}}\t" + domainNamePort + "\t" + zoneName,
		"path-only": "{{.RouteKey}}\t{{.Path}}",
	},
	objecttype.User.Name: {
		"summary": "{{.UserKey}}\t{{.LoginEmail}}",
	},
}

var preDefFormatName = ""

func init() {
	for ot, fmts := range formatStrings {
		keys := []string{}
		for k := range fmts {
			keys = append(keys, k)
		}
		preDefFormatName += fmt.Sprintf("    - %s: %s\n", ot, strings.Join(keys, ", "))
	}
}

func getFmtString(src map[string]map[string]string, ot, name string) string {
	ss, ok := src[ot]
	if !ok {
		return ""
	}
	s := ss[name]
	return s
}

func (gc *listRunner) format(objs []interface{}, otype string) error {
	fmtstr := gc.cfg.fmt
	header := ""
	if fmtstr[0] == '+' {
		fmtstr = fmtstr[1:]
		header = gc.cfg.fmtHeader
	} else {
		otype = strings.ToLower(otype)
		s := getFmtString(formatStrings, otype, fmtstr)
		header = getFmtString(headerStrings, otype, fmtstr)
		if s == "" {
			return fmt.Errorf("No available format strings for object '%s' by name of '%s'", otype, fmtstr)
		}
		fmtstr = s
	}

	t := template.New("%s")
	t = t.Funcs(
		map[string]interface{}{
			"getDomain": mkGetDomain(gc.cfg.apiClient),
			"getZone":   mkGetZone(gc.cfg.apiClient),
		},
	)

	t, err := t.Parse(fmtstr)
	if err != nil {
		return fmt.Errorf("failed to parse format string %s: %s", fmtstr, err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	if header != "" {
		fmt.Fprintln(w, header)
	}

	for _, o := range objs {
		err = t.Execute(w, o)
		if err != nil {
			return err
		}
		fmt.Fprintln(w)
	}
	w.Flush()

	return nil
}

func argsToAttrs(args []string) map[string]string {
	attr := map[string]string{}
	for _, kv := range args {
		k, v := tbnstrings.SplitFirstEqual(kv)
		attr[k] = v
	}
	return attr
}

func displayFilterFields(f interface{}, m map[string]string) {
	fmt.Printf("Listing results may be filtered by setting attributes of a %T\n", f)
	fmt.Printf("\nThe filterable attribute names and their types:\n")
	str := ""
	for k, v := range m {
		str += k + "\t" + v + "\n"
	}
	fmt.Println(
		tbnstrings.PadLeft(tbntabwriter.FormatWithHeader("NAME\tTYPE", str), 4))
}

func (gc *listRunner) run(cmd *command.Cmd, args []string) error {
	svc, err := gc.cfg.UntypedSvc(&args)
	if err != nil {
		return err
	}

	if gc.cfg.showFilterFields {
		f := svc.IndexZeroFilter()
		displayFilterFields(f, describeFields(f))
		return nil
	}

	objs, err := svc.FilteredIndex(gc.cfg.sliceSep, argsToAttrs(args))
	if err != nil {
		return err
	}

	if gc.cfg.fmt == "" {
		gc.cfg.PrintResult(objs)
		return nil
	}

	return gc.format(objs, svc.Type().Name)
}

func (gc *listRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	err := gc.run(cmd, args)
	if err != nil {
		return gc.cfg.PrettyCmdErr(cmd, err)
	}

	return command.NoError()
}

func cmdList(cfg globalConfigT) *command.Cmd {
	runner := &listRunner{&listCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "list",
		Summary:     "list all of a particular object in the Turbine Labs API",
		Usage:       "[OPTIONS] <object type> [field_name=field_value]...",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.BoolVar(&runner.cfg.showFilterFields, "show-filter-fields", false, `Displays
attributes that can be filtered when listing an object type. Each field is
specified as a single argument of the format name=value.

If an attribute expects a slice of values they may be specified as a comma-separated
string (if commas are needed in the values see {{ul "filter-slice-separator"}}). If the
attribute is a time then it should be specified as the number of milliseconds since the
unix epoch. Boolean true values may be represented by true, t, yes, y, or 1, and is case
insensitive.`)

	cmd.Flags.StringVar(
		&runner.cfg.sliceSep,
		"filter-slice-separator",
		",",
		`sets the delimiter used to indicate the boundary between elements when passing
a list of values into a filter attribute that expects a slice.`)

	cmd.Flags.StringVar(
		&runner.cfg.fmt,
		"format",
		"",
		`<format name> or +<format string>

Some pre-defined format strings may be referenced by {{ul "format name"}}; if
set this will override the more general json/yaml format flag. The available
pre-defined formats vary based on the {{ul "object type"}} being listed:

`+preDefFormatName+`

If a custom format is desired it may be specified by prefixing the string with
'+'. Custom formatting is defined using golang template syntax
(see https://golang.org/pkg/text/template). For the field reference for specific
API objects, see the API Godoc (https://godoc.org/github.com/turbinelabs/api).

{{ul "EXAMPLE"}}:

		> tbnctl --api.key=$TBN_API_KEY list \
		  --format='+{{ "{{.Name}}{{range .Instances}}" }}
		  {{ "{{.Host}}:{{.Port}}{{range .Metadata}}" }}
		    {{ "{{.Key}}={{.Value}}{{end}}{{end}}'" }} cluster

		local-demo-api-cluster
		  127.0.0.1:8080
		    stage=prod
		    version=blue
		  127.0.0.1:8081
		    stage=prod
		    version=green
		  127.0.0.1:8082
		    stage=dev
		    version=yellow
		local-demo-ui-cluster
		  127.0.0.1:8083
		    stage=prod

`,
	)

	cmd.Flags.StringVar(
		&runner.cfg.fmtHeader,
		"header",
		"",
		"Header used if a custom -format value is specified",
	)

	return cmd
}
