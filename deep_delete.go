package main

import (
	"fmt"
	"strings"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/cli/terminal"
	"github.com/turbinelabs/nonstdlib/log/console"
	tbnos "github.com/turbinelabs/nonstdlib/os"
)

type proxyMod struct {
	proxy           api.Proxy
	domainsToRemove map[api.DomainKey]api.Domain
}

type deleter struct {
	svc *unifiedSvc

	zone      api.Zone
	clusters  map[api.ClusterKey]api.Cluster
	domains   map[api.DomainKey]api.Domain
	routes    map[api.RouteKey]api.Route
	srs       map[api.SharedRulesKey]api.SharedRules
	proxies   map[api.ProxyKey]api.Proxy
	proxyMods map[api.ProxyKey]*proxyMod

	domainsForReport map[api.DomainKey]api.Domain
}

func newDeleter(svc *unifiedSvc) *deleter {
	return &deleter{
		svc:              svc,
		clusters:         make(map[api.ClusterKey]api.Cluster),
		domains:          make(map[api.DomainKey]api.Domain),
		routes:           make(map[api.RouteKey]api.Route),
		srs:              make(map[api.SharedRulesKey]api.SharedRules),
		proxies:          make(map[api.ProxyKey]api.Proxy),
		proxyMods:        make(map[api.ProxyKey]*proxyMod),
		domainsForReport: make(map[api.DomainKey]api.Domain),
	}
}

func (d *deleter) zoneIsSet() bool {
	return !d.zone.Equals(api.Zone{})
}

func (d *deleter) addZoneKey(zk api.ZoneKey) error {
	if d.zoneIsSet() {
		return fmt.Errorf("cannot delete more than one zone")
	}
	z, err := d.svc.Zone().Get(zk)
	if err != nil {
		return err
	}
	if (api.Zone{}.Equals(z)) {
		return fmt.Errorf("no zone found for key %s", zk)
	}
	d.zone = z

	clusters, err := d.svc.Cluster().Index(service.ClusterFilter{ZoneKey: zk})
	if err != nil {
		return err
	}
	for _, c := range clusters {
		d.clusters[c.ClusterKey] = c
	}

	routes, err := d.svc.Route().Index(service.RouteFilter{ZoneKey: zk})
	if err != nil {
		return err
	}
	for _, r := range routes {
		d.routes[r.RouteKey] = r
	}

	srs, err := d.svc.SharedRules().Index(service.SharedRulesFilter{ZoneKey: zk})
	if err != nil {
		return err
	}
	for _, sr := range srs {
		d.srs[sr.SharedRulesKey] = sr
	}

	domains, err := d.svc.Domain().Index(service.DomainFilter{ZoneKey: zk})
	if err != nil {
		return err
	}
	for _, dom := range domains {
		d.domains[dom.DomainKey] = dom
	}

	proxies, err := d.svc.Proxy().Index(service.ProxyFilter{ZoneKey: zk})
	if err != nil {
		return err
	}
	for _, p := range proxies {
		d.proxies[p.ProxyKey] = p
	}

	return nil
}

func (d *deleter) addDomainKey(dk api.DomainKey) error {
	if _, ok := d.domains[dk]; ok {
		return nil
	}

	dom, err := d.svc.Domain().Get(dk)
	if err != nil {
		return err
	}
	if (api.Domain{}.Equals(dom)) {
		return fmt.Errorf("no domain found for key %s", dk)
	}

	return d.addDomain(dom)
}

func (d *deleter) addDomain(dom api.Domain) error {
	dk := dom.DomainKey
	if _, ok := d.domains[dk]; ok {
		return nil
	}

	d.domains[dk] = dom

	// add any routes that may also need to be deleted
	rs, err := d.svc.Route().Index(service.RouteFilter{DomainKey: dk})
	if err != nil {
		return err
	}
	for _, r := range rs {
		if err := d.addRoute(r); err != nil {
			return err
		}
	}

	// find proxies for which the domain needs to be removed
	ps, err := d.svc.Proxy().Index(service.ProxyFilter{DomainKeys: []api.DomainKey{dk}})
	if err != nil {
		return err
	}
	for _, p := range ps {
		if _, ok := d.proxyMods[p.ProxyKey]; !ok {
			d.proxyMods[p.ProxyKey] = &proxyMod{p, map[api.DomainKey]api.Domain{dom.DomainKey: dom}}
		} else {
			d.proxyMods[p.ProxyKey].domainsToRemove[dom.DomainKey] = dom
		}
	}

	return nil
}

func (d *deleter) addRouteKey(rk api.RouteKey) error {
	if _, ok := d.routes[rk]; ok {
		return nil
	}

	r, err := d.svc.Route().Get(rk)
	if err != nil {
		return err
	}
	if (api.Route{}.Equals(r)) {
		return fmt.Errorf("no route found for key %s", rk)
	}

	return d.addRoute(r)
}

func (d *deleter) addRoute(r api.Route) error {
	rk := r.RouteKey

	if _, ok := d.routes[rk]; !ok {
		d.routes[rk] = r
	}

	if dom, ok := d.domains[r.DomainKey]; ok {
		d.domainsForReport[dom.DomainKey] = dom
		return nil
	}

	dom, err := d.svc.Domain().Get(r.DomainKey)
	if err != nil {
		return err
	}

	d.domainsForReport[dom.DomainKey] = dom
	return nil
}

func (d *deleter) addSharedRulesKey(srk api.SharedRulesKey) error {
	if _, ok := d.srs[srk]; ok {
		return nil
	}

	sr, err := d.svc.SharedRules().Get(srk)
	if err != nil {
		return err
	}
	if (api.SharedRules{}.Equals(sr)) {
		return fmt.Errorf("no shared_rules found for key %s", srk)
	}

	return d.addSharedRules(sr)
}

func (d *deleter) addSharedRules(sr api.SharedRules) error {
	srk := sr.SharedRulesKey
	if _, ok := d.srs[srk]; ok {
		return nil
	}

	d.srs[srk] = sr

	rs, err := d.svc.Route().Index(service.RouteFilter{SharedRulesKey: srk})
	if err != nil {
		return err
	}

	for _, r := range rs {
		d.addRoute(r)
	}

	return nil
}

func (d *deleter) addOrphans() error {
	// collect shared rules keys from routes
	candidates := map[api.SharedRulesKey]bool{}
	for _, r := range d.routes {
		candidates[r.SharedRulesKey] = true
	}

	if len(candidates) == 0 {
		return nil
	}

	// look up routes with those shared rules keys
	var filters []service.RouteFilter
	for srk := range candidates {
		filters = append(filters, service.RouteFilter{SharedRulesKey: srk})
	}

	rs, err := d.svc.Route().Index(filters...)
	if err != nil {
		return err
	}

	// if any of these routes is not amongst the routes to be deleted,
	// remove the SR from the candidates
	for _, r := range rs {
		if _, ok := d.routes[r.RouteKey]; !ok && candidates[r.SharedRulesKey] {
			delete(candidates, r.SharedRulesKey)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	for srk := range candidates {
		if err := d.addSharedRulesKey(srk); err != nil {
			return err
		}
	}

	return nil
}

func clusterStr(c api.Cluster) string {
	return fmt.Sprintf("Cluster(%s:%s)", c.ClusterKey, c.Name)
}

func routeStr(r api.Route, d api.Domain) string {
	return fmt.Sprintf("Route(%s:%s:%d%s)", r.RouteKey, d.Name, d.Port, r.Path)
}

func srStr(sr api.SharedRules) string {
	return fmt.Sprintf("SharedRules(%s:%s)", sr.SharedRulesKey, sr.Name)
}

func domainStr(d api.Domain) string {
	return fmt.Sprintf("Domain(%s:%s:%d)", d.DomainKey, d.Name, d.Port)
}

func proxyStr(p api.Proxy) string {
	return fmt.Sprintf("Proxy(%s:%s)", p.ProxyKey, p.Name)
}

func zoneStr(z api.Zone) string {
	return fmt.Sprintf("Zone(%s:%s)", z.ZoneKey, z.Name)
}

func proxyModStr(pm *proxyMod, indent, verb string) string {
	dNames := []string{}
	for dk, dom := range pm.domainsToRemove {
		dNames = append(dNames, fmt.Sprintf("%s:%s:%d", dk, dom.Name, dom.Port))
	}
	noun := "Domain"
	if len(dNames) > 1 {
		noun += "s"
	}
	return fmt.Sprintf(
		"%sProxy(%s:%s):\n%s  %s %s: %s",
		indent,
		pm.proxy.ProxyKey,
		pm.proxy.Name,
		indent,
		verb,
		noun,
		strings.Join(dNames, ", "),
	)
}

func (d *deleter) report() string {
	str := "Deep deletion will delete the following objects:\n"

	for _, r := range d.routes {
		dom := d.domainsForReport[r.DomainKey]
		str += "  " + routeStr(r, dom) + "\n"
	}

	for _, sr := range d.srs {
		str += "  " + srStr(sr) + "\n"
	}

	for _, p := range d.proxies {
		str += "  " + proxyStr(p) + "\n"
	}

	for _, dom := range d.domains {
		str += "  " + domainStr(dom) + "\n"
	}

	for _, c := range d.clusters {
		str += "  " + clusterStr(c) + "\n"
	}

	if d.zoneIsSet() {
		str += "  " + zoneStr(d.zone) + "\n"
	}

	if len(d.proxyMods) != 0 {
		str += "Additionally, the following proxies will be modified:\n"
		for _, pm := range d.proxyMods {
			str += proxyModStr(pm, "  ", "Remove") + "\n"
		}
	}

	str += "Proceed?"

	return str
}

func (d *deleter) execute() error {
	if !d.zoneIsSet() {
		d.addOrphans()
	}

	if ok, err := terminal.Ask(tbnos.New(), d.report()); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("canceled deep deletion")
	}

	for _, r := range d.routes {
		dom := d.domainsForReport[r.DomainKey]
		console.Info().Printf("Deleting %s", routeStr(r, dom))
		if err := d.svc.Route().Delete(r.RouteKey, r.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", routeStr(r, dom))
	}

	for _, sr := range d.srs {
		console.Info().Printf("Deleting %s", srStr(sr))
		if err := d.svc.SharedRules().Delete(sr.SharedRulesKey, sr.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", srStr(sr))
	}

	for _, p := range d.proxies {
		console.Info().Printf("Deleting %s", proxyStr(p))
		if err := d.svc.Proxy().Delete(p.ProxyKey, p.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", proxyStr(p))
	}

	for _, pm := range d.proxyMods {
		p := pm.proxy
		var dks []api.DomainKey
		for _, dk := range p.DomainKeys {
			if _, ok := pm.domainsToRemove[dk]; !ok {
				// only add back in if not to be removed
				dks = append(dks, dk)
			}
		}
		p.DomainKeys = dks

		console.Info().Println(proxyModStr(pm, "", "Deleting"))
		if _, err := d.svc.Proxy().Modify(p); err != nil {
			return err
		}
		console.Info().Println(proxyModStr(pm, "", "Deleted"))
	}

	for _, dom := range d.domains {
		console.Info().Printf("Deleting %s", domainStr(dom))
		if err := d.svc.Domain().Delete(dom.DomainKey, dom.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", domainStr(dom))
	}

	for _, c := range d.clusters {
		console.Info().Printf("Deleting %s", clusterStr(c))
		if err := d.svc.Cluster().Delete(c.ClusterKey, c.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", clusterStr(c))
	}

	if d.zoneIsSet() {
		console.Info().Printf("Deleting %s", zoneStr(d.zone))
		if err := d.svc.Zone().Delete(d.zone.ZoneKey, d.zone.Checksum); err != nil {
			return err
		}
		console.Info().Printf("Deleted %s", zoneStr(d.zone))
	}

	return nil
}

func (a clusterAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	console.Error().Println("warning: --deep ignored for cluster delete")
	return a.Delete(k, cs)
}

func (a domainAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	d := newDeleter(svc)
	if err := d.addDomainKey(api.DomainKey(k)); err != nil {
		return err
	}
	return d.execute()
}

func (a proxyAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	console.Error().Println("warning: --deep ignored for proxy delete")
	return a.Delete(k, cs)
}

func (a routeAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	d := newDeleter(svc)
	if err := d.addRouteKey(api.RouteKey(k)); err != nil {
		return err
	}
	return d.execute()
}

func (a sharedRulesAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	d := newDeleter(svc)
	if err := d.addSharedRulesKey(api.SharedRulesKey(k)); err != nil {
		return err
	}
	return d.execute()
}

func (a userAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	console.Error().Println("warning: --deep ignored for user delete")
	return a.Delete(k, cs)
}

func (a zoneAdapter) DeepDelete(k string, cs api.Checksum, svc *unifiedSvc) error {
	d := newDeleter(svc)
	if err := d.addZoneKey(api.ZoneKey(k)); err != nil {
		return err
	}
	return d.execute()
}
