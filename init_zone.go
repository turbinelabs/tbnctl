package main

import (
	"fmt"
	"strings"

	"github.com/turbinelabs/api"
	apiflags "github.com/turbinelabs/api/client/flags"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
	"github.com/turbinelabs/nonstdlib/log/console"
	tbnstrings "github.com/turbinelabs/nonstdlib/strings"
)

const (
	initZoneDesc = `
initialize a named Zone in the Turbine Labs API, adding zero or more default
routes for pairs of domain/port and cluster names, and zero or more proxies
serving one or more domains each.`

	domainFormat = "`" + `"domain:port=alias([:alias]*),..."` + "`"

	domainStrDesc = `

a comma-delimited list of pairs of a domain/port and a colon-delimited list of
aliases, of the form:

		` + domainFormat + `

Each pair is delimited by "=", both the domain/port and aliases are
colon-delimited. One can also specify multiple --domain flags, rather than
comma-delimiting. For example, when adding two domains, each with two aliases:

		--domains=foo.example.com:80=www.foo.example.com:w3.foo.example.com,bar.example.com:80=www.bar.example.com:w3.bar.example.com,

Or:

		--domains=foo.example.com:80=www.foo.example.com:w3.foo.example.com
		--domains=bar.example.com:80=www.bar.example.com:w3.bar.example.com

No routes are added, unless specified using --routes. If a domain is specified
more than once, all aliases are combined.
`

	routeFormat = "`" + `"domain:port[/path]=cluster([:key=value]*),..."` + "`"

	routeStrDesc = `
a comma-delimited list of pairs of domain/port, with an optional path, and
cluster names, with optional metadata, of the form:

    ` + routeFormat + `

Each metadatum is delimited by "=" and each metadata pair is delimited by ":".
One can also specify multiple --routes flags, rather than comma-delimiting.
For example, adding two domains, one with two routes and metadata on one cluster:

    --routes=example.com:80=exampleService,api.example.com:443=apiService,api.example.com:443/users=userService:stage=prod:version=1.0

Or:

    --routes=example.com:80=exampleService \
    --routes=api.example.com:443=apiService \
    --routes=api.example.com:443/users=userService:stage=prod:version=1.0

If no corresponding domain is specified using --domains, a domain is added. More
than one specification of the same route is considered an error.
`
	proxyFmt = "`" + `"proxy=domain:port,..."` + "`"

	proxyStrDesc = `
a comma-delimited list of pairs of proxy and domain/port, of the form:

    ` + proxyFmt + `

In order to serve more than one domain/port from a given proxy, you may specify
the same proxy more than once, with a different domain/port.

One can also specify multiple --proxy flags, rather than comma-delimiting.
`
)

func cmdInitZone(globalConfig globalConfigT) *command.Cmd {
	cmd := &command.Cmd{
		Name:        "init-zone",
		Summary:     "initialize a named Zone in the Turbine Labs API",
		Usage:       "[OPTIONS] <zone-name> ",
		Description: initZoneDesc,
	}

	r := &initZoneRunner{
		apiFlags:   globalConfig.apiFlags,
		domainStrs: tbnflag.NewStrings(),
		routesStrs: tbnflag.NewStrings(),
		proxyStrs:  tbnflag.NewStrings(),
	}

	cmd.Flags.Var(&r.domainStrs, "domains", domainStrDesc)
	cmd.Flags.Var(&r.routesStrs, "routes", routeStrDesc)
	cmd.Flags.Var(&r.proxyStrs, "proxies", proxyStrDesc)
	cmd.Flags.BoolVar(
		&r.replace,
		"replace",
		false,
		"If true, replace existing Routes, SharedRules, and Proxies. If false, leave them as is.",
	)

	cmd.Runner = r
	return cmd
}

type initZoneRunner struct {
	domainStrs tbnflag.Strings
	routesStrs tbnflag.Strings
	proxyStrs  tbnflag.Strings
	apiFlags   apiflags.ClientFromFlags
	replace    bool
}

type hostPort struct {
	host string
	port int
}

func (hp hostPort) String() string {
	return fmt.Sprintf("%s:%d", hp.host, hp.port)
}

type domain struct {
	name    hostPort
	aliases api.DomainAliases
}

type proxy struct {
	name    string
	domains []hostPort
}

type route struct {
	domain   hostPort
	cluster  string
	path     string
	metadata api.Metadata
}

func (r route) String() string {
	return fmt.Sprintf("%s%s", r.domain, r.path)
}

func (r *initZoneRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if len(args) != 1 {
		return cmd.BadInput("requires exactly one argument")
	}
	zoneName := args[0]

	domains, err := parseDomains(r.domainStrs.Strings)
	if err != nil {
		return cmd.BadInput(err)
	}

	routes, err := parseRoutes(r.routesStrs.Strings)
	if err != nil {
		return cmd.BadInput(err)
	}

	if err := validateRoutes(routes); err != nil {
		return cmd.BadInput(err)
	}

	proxies, err := parseProxies(r.proxyStrs.Strings)
	if err != nil {
		return cmd.BadInput(err)
	}

	if err := validateProxies(proxies, routes); err != nil {
		return cmd.BadInput(err)
	}

	if err := r.apiFlags.Validate(); err != nil {
		return cmd.BadInput(err)
	}

	console.Debug().Println("ZONE NAME: ", zoneName)
	console.Debug().Printf("  DOMAINS: %+v\n", domains)
	console.Debug().Printf("   ROUTES: %+v\n", routes)
	console.Debug().Printf("  PROXIES: %+v\n", proxies)

	svc, err := r.apiFlags.Make()
	if err != nil {
		return cmd.Error(err)
	}

	zkey, err := addZone(svc.Zone(), zoneName)
	if err != nil {
		return cmd.Error(err)
	}

	if err := addObjects(svc, zkey, domains, routes, proxies, r.replace); err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
}

func parseDomains(strs []string) ([]domain, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	idxMap := map[hostPort]int{}
	result := []domain{}
	for _, str := range strs {
		d, err := parseDomain(str)
		if err != nil {
			return nil, err
		}
		if idx, ok := idxMap[d.name]; ok {
			result[idx].aliases = append(result[idx].aliases, d.aliases...)
		} else {
			idxMap[d.name] = len(result)
			result = append(result, d)
		}
	}
	return result, nil
}

func parseDomain(str string) (domain, error) {
	domainHostPort, aliasesStr := tbnstrings.SplitFirstEqual(str)
	if domainHostPort == "" || aliasesStr == "" {
		return domain{}, fmt.Errorf("malformed domains argument: %q", str)
	}

	dom, port, err := tbnstrings.SplitHostPort(domainHostPort)
	if err != nil {
		return domain{}, fmt.Errorf(
			"malformed domain/port %q in domains argument %q",
			domainHostPort,
			str,
		)
	}

	aliasesStrs := strings.Split(aliasesStr, ":")
	aliases := make(api.DomainAliases, len(aliasesStrs), len(aliasesStrs))
	for i := range aliasesStrs {
		aliases[i] = api.DomainAlias(aliasesStrs[i])
	}

	return domain{name: hostPort{dom, port}, aliases: aliases}, nil
}

func parseProxies(strs []string) ([]proxy, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	idxMap := map[string]int{}
	result := []proxy{}
	for _, str := range strs {
		p, err := parseProxy(str)
		if err != nil {
			return nil, err
		}
		if idx, ok := idxMap[p.name]; ok {
			result[idx].domains = append(result[idx].domains, p.domains...)
		} else {
			idxMap[p.name] = len(result)
			result = append(result, p)
		}
	}
	return result, nil
}

func parseProxy(str string) (proxy, error) {
	proxyName, domainHostPort := tbnstrings.SplitFirstEqual(str)
	if proxyName == "" || domainHostPort == "" {
		return proxy{}, fmt.Errorf("malformed proxy argument: %q", str)
	}

	domain, port, err := tbnstrings.SplitHostPort(domainHostPort)
	if err != nil {
		return proxy{}, fmt.Errorf("malformed domain/port %q in proxy argument %q", domainHostPort, str)
	}

	return proxy{proxyName, []hostPort{{domain, port}}}, nil
}

func validateProxies(proxies []proxy, routes []route) error {
	dMap := map[string]bool{}
	for _, r := range routes {
		dMap[r.domain.String()] = true
	}
	for _, p := range proxies {
		for _, d := range p.domains {
			if !dMap[d.String()] {
				return fmt.Errorf("proxy %s refers to unknown domain %s", p.name, d)
			}
		}
	}
	return nil
}

func parseRoutes(strs []string) ([]route, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	seen := map[string]bool{}
	result := make([]route, 0, len(strs))
	for _, str := range strs {
		r, err := parseRoute(str)
		if err != nil {
			return nil, err
		}
		if seen[r.String()] {
			return nil, fmt.Errorf("Route %q specified more than once", r)
		}
		seen[r.String()] = true
		result = append(result, r)
	}
	return result, nil
}

func parseRoute(str string) (route, error) {
	domainHostPortPath, clusterAndMetaDef := tbnstrings.SplitFirstEqual(str)
	if domainHostPortPath == "" || clusterAndMetaDef == "" {
		return route{}, fmt.Errorf("malformed route %q", str)
	}

	domainHostPort, pathDef := tbnstrings.Split2(domainHostPortPath, "/")

	domain, port, err := tbnstrings.SplitHostPort(domainHostPort)
	if err != nil {
		return route{}, fmt.Errorf("malformed domain/port %q in route argument %q", domainHostPort, str)
	}

	path := "/"
	if pathDef != "" {
		path = "/" + pathDef
	}

	clusterAndMeta := strings.Split(clusterAndMetaDef, ":")
	cluster := clusterAndMeta[0]
	if cluster == "" {
		return route{}, fmt.Errorf("empty cluster name in route argument %q", str)
	}

	metadata := api.Metadata{}
	for _, kv := range clusterAndMeta[1:] {
		key, value := tbnstrings.SplitFirstEqual(kv)
		if key == "" {
			return route{}, fmt.Errorf("malformed metadata %q in route argument %q", kv, str)
		}
		metadata = append(metadata, api.Metadatum{key, value})
	}

	return route{hostPort{domain, port}, cluster, path, metadata}, nil
}

func validateRoutes(routes []route) error {
	cMap := map[string]bool{}
	for _, r := range routes {
		key := r.domain.String() + r.path
		if cMap[key] {
			return fmt.Errorf("route %s declared more than once", key)
		}
		cMap[key] = true
	}
	return nil
}

func addZone(zoneSvc service.Zone, name string) (api.ZoneKey, error) {
	zones, err := zoneSvc.Index(service.ZoneFilter{Name: name})
	if err != nil {
		return "", err
	}

	if len(zones) > 0 {
		console.Debug().Printf("Zone %s already exists\n", name)
		return zones[0].ZoneKey, nil
	}

	console.Debug().Printf("Creating Zone %s\n", name)
	zone, err := zoneSvc.Create(api.Zone{Name: name})
	if err != nil {
		return "", err
	}

	console.Debug().Printf("Created Zone %s\n", name)
	return zone.ZoneKey, nil
}

func addDomain(
	svc service.Domain,
	zkey api.ZoneKey,
	dom domain,
	replace bool,
) (api.DomainKey, error) {
	ds, err := svc.Index(service.DomainFilter{Name: dom.name.host, ZoneKey: zkey})
	if err != nil {
		return "", err
	}
	domain := api.Domain{}
	for _, d := range ds {
		if d.Port == dom.name.port {
			domain = d
			break
		}
	}

	if domain.DomainKey != "" {
		console.Debug().Printf("Domain %s already exists\n", dom.name)
		if replace && !domain.Aliases.Equals(dom.aliases) {
			console.Debug().Printf("Aliases differ, modifying Domain %s\n", dom.name)
			domain.Aliases = dom.aliases
			domain, err = svc.Modify(domain)
			if err != nil {
				return "", err
			}
			console.Debug().Printf("Modified Domain %s\n", dom.name)
		}
		return domain.DomainKey, nil
	}

	console.Debug().Printf("Creating Domain %s\n", dom.name)
	domain, err = svc.Create(api.Domain{
		Name:    dom.name.host,
		Port:    dom.name.port,
		Aliases: dom.aliases,
		ZoneKey: zkey,
	})
	if err != nil {
		return "", err
	}
	console.Debug().Printf("Created Domain %s\n", dom.name)
	return domain.DomainKey, nil
}

func addCluster(svc service.Cluster, zkey api.ZoneKey, name string) (api.ClusterKey, error) {
	cs, err := svc.Index(service.ClusterFilter{Name: name, ZoneKey: zkey})
	if err != nil {
		return "", err
	}

	if len(cs) > 0 {
		console.Debug().Printf("Cluster %s already exists\n", name)
		return cs[0].ClusterKey, nil
	}

	console.Debug().Printf("Creating Cluster %s\n", name)
	cluster, err := svc.Create(api.Cluster{Name: name, ZoneKey: zkey})
	if err != nil {
		return "", err
	}
	console.Debug().Printf("Created Cluster %s\n", name)
	return cluster.ClusterKey, nil
}

func addSharedRules(
	svc service.SharedRules,
	name string,
	zkey api.ZoneKey,
	ckey api.ClusterKey,
	metadata api.Metadata,
	replace bool,
) (api.SharedRulesKey, error) {
	srs, err := svc.Index(service.SharedRulesFilter{Name: name, ZoneKey: zkey})
	if err != nil {
		return "", err
	}

	sr := api.SharedRules{
		Name:    name,
		ZoneKey: zkey,
		Default: api.AllConstraints{
			Light: api.ClusterConstraints{
				{
					ClusterKey: ckey,
					Metadata:   metadata,
					Weight:     1,
				},
			},
		},
	}

	if len(srs) > 0 {
		console.Debug().Printf("SharedRules %s already exists\n", name)
		if !replace {
			return srs[0].SharedRulesKey, nil
		}

		sr.Checksum = srs[0].Checksum
		sr.SharedRulesKey = srs[0].SharedRulesKey

		console.Debug().Printf("Modifying SharedRules %s\n", name)
		if sr, err = svc.Modify(sr); err != nil {
			return "", err
		}
		console.Debug().Printf("Modified SharedRules %s\n", name)
	} else {
		console.Debug().Printf("Creating SharedRules %s\n", name)
		if sr, err = svc.Create(sr); err != nil {
			return "", err
		}
		console.Debug().Printf("Created SharedRules %s\n", name)
	}

	return sr.SharedRulesKey, nil
}

// first pass to create domains, clusters, and shared rules
func addRouteDependencies(
	svc service.All,
	zkey api.ZoneKey,
	domains []domain,
	routes []route,
	replace bool,
) (
	map[string]api.SharedRulesKey,
	map[string]api.DomainKey,
	error,
) {
	sharedRulesKeyMap := map[string]api.SharedRulesKey{}
	domainKeyMap := map[string]api.DomainKey{}

	for _, d := range domains {
		dkey, err := addDomain(svc.Domain(), zkey, d, replace)
		if err != nil {
			return nil, nil, err
		}
		domainKeyMap[d.name.String()] = dkey
	}

	for _, r := range routes {
		name := r.cluster

		if _, ok := domainKeyMap[r.domain.String()]; !ok {
			dkey, err := addDomain(svc.Domain(), zkey, domain{name: r.domain}, false)
			if err != nil {
				return nil, nil, err
			}
			domainKeyMap[r.domain.String()] = dkey
		}

		ckey, err := addCluster(svc.Cluster(), zkey, name)
		if err != nil {
			return nil, nil, err
		}

		srkey, err := addSharedRules(svc.SharedRules(), name, zkey, ckey, r.metadata, replace)
		if err != nil {
			return nil, nil, err
		}
		sharedRulesKeyMap[name] = srkey
	}

	return sharedRulesKeyMap, domainKeyMap, nil
}

// second pass to create routes
func addRoutes(
	svc service.Route,
	zkey api.ZoneKey,
	routes []route,
	sharedRulesKeyMap map[string]api.SharedRulesKey,
	domainKeyMap map[string]api.DomainKey,
	replace bool,
) error {
	for _, r := range routes {
		dkey := domainKeyMap[r.domain.String()]
		srkey := sharedRulesKeyMap[r.cluster]

		rs, err := svc.Index(service.RouteFilter{DomainKey: dkey, Path: r.path, ZoneKey: zkey})
		if err != nil {
			return err
		}

		route := api.Route{DomainKey: dkey, ZoneKey: zkey, Path: r.path, SharedRulesKey: srkey}

		if len(rs) > 0 {
			console.Debug().Printf("Route already exists for %s%s\n", r.domain, r.path)
			if !replace {
				continue
			}

			route.Checksum = rs[0].Checksum
			route.RouteKey = rs[0].RouteKey

			console.Debug().Printf("Modifying Route for %s%s to %s\n", r.domain, r.path, r.cluster)
			if _, err = svc.Modify(route); err != nil {
				return err
			}
			console.Debug().Printf("Modified Route for %s%s to %s\n", r.domain, r.path, r.cluster)
		} else {
			console.Debug().Printf("Creating Route for %s%s to %s\n", r.domain, r.path, r.cluster)
			if _, err = svc.Create(route); err != nil {
				return err
			}
			console.Debug().Printf("Created Route for %s%s to %s\n", r.domain, r.path, r.cluster)
		}
	}

	return nil
}

func addProxies(
	svc service.Proxy,
	zkey api.ZoneKey,
	proxies []proxy,
	domainKeyMap map[string]api.DomainKey,
	replace bool,
) error {
	for _, p := range proxies {
		dkeys := []api.DomainKey{}
		for _, d := range p.domains {
			if dkey, ok := domainKeyMap[d.String()]; ok {
				dkeys = append(dkeys, dkey)
			} else {
				console.Debug().Printf("Ignoring unknown domain %s for proxy %s\n", d, p.name)
			}
		}

		ps, err := svc.Index(service.ProxyFilter{Name: p.name, ZoneKey: zkey})
		if err != nil {
			return err
		}

		proxy := api.Proxy{Name: p.name, DomainKeys: dkeys, ZoneKey: zkey}

		if len(ps) > 0 {
			console.Debug().Printf("Proxy %s already exists\n", p.name)
			if !replace {
				continue
			}

			proxy.Checksum = ps[0].Checksum
			proxy.ProxyKey = ps[0].ProxyKey

			console.Debug().Printf("Modifying Proxy %s\n", p.name)
			if _, err = svc.Modify(proxy); err != nil {
				return err
			}
			console.Debug().Printf("Modified Proxy %s\n", p.name)
		} else {
			console.Debug().Printf("Creating Proxy %s\n", p.name)
			if _, err = svc.Create(proxy); err != nil {
				return err
			}
			console.Debug().Printf("Created Proxy %s\n", p.name)
		}
	}
	return nil
}

func addObjects(
	svc service.All,
	zkey api.ZoneKey,
	domains []domain,
	routes []route,
	proxies []proxy,
	replace bool,
) error {
	// first pass to create domains, clusters, and shared rules
	sharedRulesKeyMap, domainKeyMap, err := addRouteDependencies(svc, zkey, domains, routes, replace)
	if err != nil {
		return err
	}

	// second pass to create routes
	if err := addRoutes(
		svc.Route(),
		zkey,
		routes,
		sharedRulesKeyMap,
		domainKeyMap,
		replace,
	); err != nil {
		return err
	}

	// third pass to create proxies
	if err := addProxies(svc.Proxy(), zkey, proxies, domainKeyMap, replace); err != nil {
		return err
	}

	return nil
}
