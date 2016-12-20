package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/cli/command"
)

type listCfg struct {
	*globalConfigT

	fmt       string
	fmtHeader string
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
		preDefFormatName += fmt.Sprintf("{{bold \"%s\"}}: %s\n\n", ot, strings.Join(keys, ", "))
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

func (gc *listRunner) run(cmd *command.Cmd, args []string) error {
	svc, err := gc.cfg.UntypedSvc(args)
	if err != nil {
		return err
	}

	objs, err := svc.Index()
	if gc.cfg.fmt == "" {
		return gc.cfg.MkResult(objs, err)
	}

	if err != nil {
		return err
	}

	return gc.format(objs, args[0])
}

func (gc *listRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if err := gc.cfg.Prepare(cmd); err != command.NoError() {
		return err
	}

	err := gc.run(cmd, args)
	if err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

func cmdList(cfg globalConfigT) *command.Cmd {
	runner := &listRunner{&listCfg{}}
	runner.cfg.globalConfigT = &cfg

	cmd := &command.Cmd{
		Name:        "list",
		Summary:     "list all of a particular object in the Turbine Labs API",
		Usage:       "[OPTIONS] <object type>",
		Description: "object type is one of: " + objTypeNames(),
		Runner:      runner,
	}

	cmd.Flags.StringVar(
		&runner.cfg.fmt,
		"format",
		"",
		`<format name> or +<format string>

Some pre-defined format strings may be refrenceed by {{ul "format name"}}; if
set this will override the more general json/yaml format flag. If a custom
format is desired it may be specified by prefixing the string with '+'. The
available pre-defined formats vary based on the {{ul "object type"}} being
listed:

`+preDefFormatName,
	)

	cmd.Flags.StringVar(
		&runner.cfg.fmtHeader,
		"header",
		"",
		"Header used if a custom -format value is specified",
	)

	return cmd
}
