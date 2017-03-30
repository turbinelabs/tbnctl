package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/turbinelabs/api"
	apiflags "github.com/turbinelabs/api/client/flags"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
	"github.com/turbinelabs/nonstdlib/log/console"
)

const (
	initZoneDesc = `
initialize a named Zone in the Turbine Labs API, adding zero or more default
routes for pairs of domain/port and cluster names, and zero or more proxies
serving one or more domains each.`

	routeFormat = "`" + `"domain:port[/path]=cluster([:key=value]*),..."` + "`"

	routeStrDesc = `
a comma-delimited list of pairs of domain/port, with an optional path, and
cluster names, with optional metadata, of the form:

    ` + routeFormat + `

Each metadatum is delimited by "=" and each metadata pair is delimited by ":".
One can also specify multiple --route flags, rather than comma-delimiting.
For example, adding two domains, one with two routes and metadata on one cluster:

    --routes=example.com:80=exampleService,api.example.com:443=apiService,api.example.com:443/users=userService:stage=prod:version=1.0

Or:

    --routes=example.com:80=exampleService \
    --routes=api.example.com:443=apiService \
    --routes=api.example.com:443/users=userService:stage=prod:version=1.0
`
	proxyFmt = "`" + `"proxy=domain:port,..."` + "`"

	proxyStrDesc = `
a comma-delimited list of pairs of proxy and domain/port, of the form:

    ` + proxyFmt + `

In order to serve more than one domain/port from a given proxy, you may specify
the same proxy more than once, with a different domain/port.
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
		routesStrs: tbnflag.NewStrings(),
		proxyStrs:  tbnflag.NewStrings(),
	}

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

func (r *initZoneRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if len(args) != 1 {
		return cmd.BadInput("requires exactly one argument")
	}
	zoneName := args[0]

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

	console.Error().Println("ZONE NAME: ", zoneName)
	console.Error().Printf("ROUTES: %+v\n", routes)
	console.Error().Printf("PROXIES: %+v\n", proxies)

	svc, err := r.apiFlags.Make()
	if err != nil {
		return cmd.Error(err)
	}

	zkey, err := addZone(svc.Zone(), zoneName)
	if err != nil {
		return cmd.Error(err)
	}

	if err := addRoutesAndProxies(svc, zkey, routes, proxies, r.replace); err != nil {
		return cmd.Error(err)
	}

	return command.NoError()
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
	pair := strings.Split(str, "=")
	if len(pair) != 2 {
		return proxy{}, fmt.Errorf("malformed proxy argument: %q", str)
	}
	proxyName := pair[0]

	domainPair := strings.Split(pair[1], ":")
	if len(domainPair) != 2 {
		return proxy{}, fmt.Errorf("malformed domain/port %q in proxy argument %q", pair[0], str)
	}
	domain := domainPair[0]

	port, err := strconv.Atoi(domainPair[1])
	if err != nil {
		return proxy{}, fmt.Errorf("malformed port %q in proxy argument %q", domainPair[1], str)
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
	result := make([]route, 0, len(strs))
	for _, str := range strs {
		r, err := parseRoute(str)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func parseRoute(str string) (route, error) {
	pair := strings.SplitN(str, "=", 2)
	if len(pair) != 2 {
		return route{}, fmt.Errorf("malformed route %q", str)
	}

	domainPair := strings.Split(pair[0], ":")
	if len(domainPair) != 2 {
		return route{}, fmt.Errorf("malformed domain/port %q in route argument %q", pair[0], str)
	}
	domain := domainPair[0]

	portPathPair := strings.SplitN(domainPair[1], "/", 2)
	port, err := strconv.Atoi(portPathPair[0])
	if err != nil {
		return route{}, fmt.Errorf("malformed port %q in route argument %q", domainPair[1], str)
	}

	path := "/"
	if len(portPathPair) == 2 {
		path = "/" + portPathPair[1]
	}

	clusterAndMeta := strings.Split(pair[1], ":")
	cluster := clusterAndMeta[0]
	if cluster == "" {
		return route{}, fmt.Errorf("empty cluster name in route argument %q", str)
	}

	metadata := api.Metadata{}
	for _, kv := range clusterAndMeta[1:] {
		kvPair := strings.SplitN(kv, "=", 2)
		if len(kvPair) != 2 {
			return route{}, fmt.Errorf("malformed metadata %q in route argument %q", kv, str)
		}
		metadata = append(metadata, api.Metadatum{kvPair[0], kvPair[1]})
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
		console.Error().Printf("Zone %s already exists\n", name)
		return zones[0].ZoneKey, nil
	}

	console.Error().Printf("Creating Zone %s\n", name)
	zone, err := zoneSvc.Create(api.Zone{Name: name})
	if err != nil {
		return "", err
	}

	console.Error().Printf("Created Zone %s\n", name)
	return zone.ZoneKey, nil
}

func addDomain(svc service.Domain, zkey api.ZoneKey, hp hostPort) (api.DomainKey, error) {
	ds, err := svc.Index(service.DomainFilter{Name: hp.host, ZoneKey: zkey})
	if err != nil {
		return "", err
	}
	domain := api.Domain{}
	for _, d := range ds {
		if d.Port == hp.port {
			domain = d
			break
		}
	}

	if domain.DomainKey != "" {
		console.Error().Printf("Domain %s already exists\n", hp)
		return domain.DomainKey, nil
	}

	console.Error().Printf("Creating Domain %s\n", hp)
	domain, err = svc.Create(api.Domain{Name: hp.host, Port: hp.port, ZoneKey: zkey})
	if err != nil {
		return "", err
	}
	console.Error().Printf("Created Domain %s\n", hp)
	return domain.DomainKey, nil
}

func addCluster(svc service.Cluster, zkey api.ZoneKey, name string) (api.ClusterKey, error) {
	cs, err := svc.Index(service.ClusterFilter{Name: name, ZoneKey: zkey})
	if err != nil {
		return "", err
	}

	if len(cs) > 0 {
		console.Error().Printf("Cluster %s already exists\n", name)
		return cs[0].ClusterKey, nil
	}

	console.Error().Printf("Creating Cluster %s\n", name)
	cluster, err := svc.Create(api.Cluster{Name: name, ZoneKey: zkey})
	if err != nil {
		return "", err
	}
	console.Error().Printf("Created Cluster %s\n", name)
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
		console.Error().Printf("SharedRules %s already exists\n", name)
		if !replace {
			return srs[0].SharedRulesKey, nil
		}

		sr.Checksum = srs[0].Checksum
		sr.SharedRulesKey = srs[0].SharedRulesKey

		console.Error().Printf("Modifying SharedRules %s\n", name)
		if sr, err = svc.Modify(sr); err != nil {
			return "", err
		}
		console.Error().Printf("Modified SharedRules %s\n", name)
	} else {
		console.Error().Printf("Creating SharedRules %s\n", name)
		if sr, err = svc.Create(sr); err != nil {
			return "", err
		}
		console.Error().Printf("Created SharedRules %s\n", name)
	}

	return sr.SharedRulesKey, nil
}

// first pass to create domains, clusters, and shared rules
func addRouteDependencies(
	svc service.All,
	zkey api.ZoneKey,
	routes []route,
	replace bool,
) (
	map[string]api.SharedRulesKey,
	map[string]api.DomainKey,
	error,
) {
	sharedRulesKeyMap := map[string]api.SharedRulesKey{}
	domainKeyMap := map[string]api.DomainKey{}

	for _, r := range routes {
		name := r.cluster

		dkey, err := addDomain(svc.Domain(), zkey, r.domain)
		if err != nil {
			return nil, nil, err
		}
		domainKeyMap[r.domain.String()] = dkey

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
			console.Error().Printf("Route already exists for %s%s\n", r.domain, r.path)
			if !replace {
				continue
			}

			route.Checksum = rs[0].Checksum
			route.RouteKey = rs[0].RouteKey

			console.Error().Printf("Modifying Route for %s%s to %s\n", r.domain, r.path, r.cluster)
			if _, err = svc.Modify(route); err != nil {
				return err
			}
			console.Error().Printf("Modified Route for %s%s to %s\n", r.domain, r.path, r.cluster)
		} else {
			console.Error().Printf("Creating Route for %s%s to %s\n", r.domain, r.path, r.cluster)
			if _, err = svc.Create(route); err != nil {
				return err
			}
			console.Error().Printf("Created Route for %s%s to %s\n", r.domain, r.path, r.cluster)
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
				console.Error().Printf("Ignoring unknown domain %s for proxy %s\n", d, p.name)
			}
		}

		ps, err := svc.Index(service.ProxyFilter{Name: p.name, ZoneKey: zkey})
		if err != nil {
			return err
		}

		proxy := api.Proxy{Name: p.name, DomainKeys: dkeys, ZoneKey: zkey}

		if len(ps) > 0 {
			console.Error().Printf("Proxy %s already exists\n", p.name)
			if !replace {
				continue
			}

			proxy.Checksum = ps[0].Checksum
			proxy.ProxyKey = ps[0].ProxyKey

			console.Error().Printf("Modifying Proxy %s\n", p.name)
			if _, err = svc.Modify(proxy); err != nil {
				return err
			}
			console.Error().Printf("Modified Proxy %s\n", p.name)
		} else {
			console.Error().Printf("Creating Proxy %s\n", p.name)
			if _, err = svc.Create(proxy); err != nil {
				return err
			}
			console.Error().Printf("Created Proxy %s\n", p.name)
		}
	}
	return nil
}

func addRoutesAndProxies(
	svc service.All,
	zkey api.ZoneKey,
	routes []route,
	proxies []proxy,
	replace bool,
) error {
	// first pass to create domains, clusters, and shared rules
	sharedRulesKeyMap, domainKeyMap, err := addRouteDependencies(svc, zkey, routes, replace)
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
