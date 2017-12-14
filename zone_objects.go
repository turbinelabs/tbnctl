package main

import (
	"fmt"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/codec"
)

// zoneObjects holds all objects for a zone, plus some keymaps to maintain
// referential integrety
type zoneObjects struct {
	Zone        api.Zone             `json:"zone"`
	Clusters    api.Clusters         `json:"clusters"`
	Domains     api.Domains          `json:"domains"`
	Proxies     api.Proxies          `json:"proxies"`
	Routes      api.Routes           `json:"routes"`
	SharedRules api.SharedRulesSlice `json:"shared_rules"`

	clusterKeyMap     map[api.ClusterKey]api.ClusterKey
	domainKeyMap      map[api.DomainKey]api.DomainKey
	sharedRulesKeyMap map[api.SharedRulesKey]api.SharedRulesKey
}

func (zo *zoneObjects) nameifyRules(rs api.Rules) {
	for i := range rs {
		rs[i].RuleKey = ""
		zo.nameifyAllConstraints(&rs[i].Constraints)
	}
}

func (zo *zoneObjects) nameifyAllConstraints(ac *api.AllConstraints) {
	zo.nameifyClusterConstraints(ac.Light)
	zo.nameifyClusterConstraints(ac.Dark)
	zo.nameifyClusterConstraints(ac.Tap)
}

func (zo *zoneObjects) nameifyClusterConstraints(ccs api.ClusterConstraints) {
	for i := range ccs {
		ccs[i].ClusterKey = zo.clusterKeyMap[ccs[i].ClusterKey]
	}
}

func (zo *zoneObjects) denameifyRules(rs api.Rules) {
	for i := range rs {
		zo.denameifyAllConstraints(&rs[i].Constraints)
	}
}

func (zo *zoneObjects) denameifyAllConstraints(ac *api.AllConstraints) {
	zo.denameifyClusterConstraints(ac.Light)
	zo.denameifyClusterConstraints(ac.Dark)
	zo.denameifyClusterConstraints(ac.Tap)
}

func (zo *zoneObjects) denameifyClusterConstraints(ccs api.ClusterConstraints) {
	for i := range ccs {
		ccs[i].ClusterKey = zo.clusterKeyMap[ccs[i].ClusterKey]
		ccs[i].ConstraintKey = ""
	}
}

func newZoneObjects() *zoneObjects {
	return &zoneObjects{
		clusterKeyMap:     map[api.ClusterKey]api.ClusterKey{},
		domainKeyMap:      map[api.DomainKey]api.DomainKey{},
		sharedRulesKeyMap: map[api.SharedRulesKey]api.SharedRulesKey{},
	}
}

// exportZone exports the zone with a ZoneKey or Name matching the given string,
// and with object keys replaced by human-readable names.
func exportZone(svc service.All, keyOrName string) (*zoneObjects, error) {
	var z api.Zone

	zs, err := svc.Zone().Index(service.ZoneFilter{Name: keyOrName})
	if err != nil {
		return nil, err
	}

	if len(zs) == 1 {
		z = zs[0]
	} else {
		z, err = svc.Zone().Get(api.ZoneKey(keyOrName))
		if err != nil {
			return nil, err
		}
	}

	zk := z.ZoneKey
	z.ZoneKey = api.ZoneKey(z.Name)
	z.Checksum = api.Checksum{}

	zo := newZoneObjects()
	zo.Zone = z

	cs, err := svc.Cluster().Index(service.ClusterFilter{ZoneKey: zk})
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		ck := api.ClusterKey(c.Name)
		zo.clusterKeyMap[c.ClusterKey] = ck
		c.ZoneKey = zo.Zone.ZoneKey
		c.ClusterKey = ck
		c.Instances = nil
		c.Checksum = api.Checksum{}
		zo.Clusters = append(zo.Clusters, c)
	}

	ds, err := svc.Domain().Index(service.DomainFilter{ZoneKey: zk})
	if err != nil {
		return nil, err
	}
	for _, d := range ds {
		dk := api.DomainKey(d.Addr())
		zo.domainKeyMap[d.DomainKey] = dk
		d.ZoneKey = zo.Zone.ZoneKey
		d.DomainKey = dk
		d.Checksum = api.Checksum{}
		zo.Domains = append(zo.Domains, d)
	}

	ps, err := svc.Proxy().Index(service.ProxyFilter{ZoneKey: zk})
	if err != nil {
		return nil, err
	}
	for _, p := range ps {
		p.ZoneKey = zo.Zone.ZoneKey
		dks := make([]api.DomainKey, len(p.DomainKeys), len(p.DomainKeys))
		for i, dk := range p.DomainKeys {
			dks[i] = zo.domainKeyMap[dk]
		}
		p.DomainKeys = dks
		p.ProxyKey = api.ProxyKey(p.Name)
		p.Checksum = api.Checksum{}
		zo.Proxies = append(zo.Proxies, p)
	}

	srs, err := svc.SharedRules().Index(service.SharedRulesFilter{ZoneKey: zk})
	if err != nil {
		return nil, err
	}
	for _, sr := range srs {
		srk := api.SharedRulesKey(sr.Name)
		zo.sharedRulesKeyMap[sr.SharedRulesKey] = srk
		sr.ZoneKey = zo.Zone.ZoneKey
		sr.SharedRulesKey = srk
		sr.Checksum = api.Checksum{}
		zo.nameifyAllConstraints(&sr.Default)
		zo.nameifyRules(sr.Rules)
		zo.SharedRules = append(zo.SharedRules, sr)
	}

	rs, err := svc.Route().Index(service.RouteFilter{ZoneKey: zk})
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		r.ZoneKey = zo.Zone.ZoneKey
		r.DomainKey = zo.domainKeyMap[r.DomainKey]
		r.SharedRulesKey = zo.sharedRulesKeyMap[r.SharedRulesKey]
		r.RouteKey = api.RouteKey(fmt.Sprintf("%s%s", r.DomainKey, r.Path))
		r.Checksum = api.Checksum{}
		zo.nameifyRules(r.Rules)
		zo.Routes = append(zo.Routes, r)
	}

	return zo, nil
}

// importZone decodes the zone encoded in the txt using the provided Codec,
// and then stores it in the API using the given name. Objects are stored in
// dependency order, to maintain referential integrity.
func importZone(svc service.All, name string, cdc codec.Codec, txt string) (*zoneObjects, error) {
	zo := newZoneObjects()
	if err := codec.DecodeFromString(cdc, txt, zo); err != nil {
		return nil, err
	}

	var err error
	zo.Zone.Name = name
	zo.Zone.ZoneKey = ""
	zo.Zone, err = svc.Zone().Create(zo.Zone)
	if err != nil {
		return nil, err
	}

	for i := range zo.Clusters {
		c := zo.Clusters[i]
		c.ZoneKey = zo.Zone.ZoneKey
		c.ClusterKey = ""
		c, err = svc.Cluster().Create(c)
		if err != nil {
			return nil, err
		}
		zo.clusterKeyMap[api.ClusterKey(c.Name)] = c.ClusterKey
		zo.Clusters[i] = c
	}

	for i := range zo.Domains {
		d := zo.Domains[i]
		d.ZoneKey = zo.Zone.ZoneKey
		d.DomainKey = ""
		d, err = svc.Domain().Create(d)
		if err != nil {
			return nil, err
		}
		zo.domainKeyMap[api.DomainKey(d.Addr())] = d.DomainKey
		zo.Domains[i] = d
	}

	for i := range zo.Proxies {
		p := zo.Proxies[i]
		p.ZoneKey = zo.Zone.ZoneKey
		p.ProxyKey = ""
		length := len(p.DomainKeys)
		dks := make([]api.DomainKey, length, length)
		for i, dk := range p.DomainKeys {
			dks[i] = zo.domainKeyMap[dk]
		}
		p.DomainKeys = dks
		p, err = svc.Proxy().Create(p)
		if err != nil {
			return nil, err
		}
		zo.Proxies[i] = p
	}

	for i := range zo.SharedRules {
		sr := zo.SharedRules[i]
		sr.ZoneKey = zo.Zone.ZoneKey
		sr.SharedRulesKey = ""
		zo.denameifyAllConstraints(&sr.Default)
		zo.denameifyRules(sr.Rules)
		sr, err = svc.SharedRules().Create(sr)
		if err != nil {
			return nil, err
		}
		zo.sharedRulesKeyMap[api.SharedRulesKey(sr.Name)] = sr.SharedRulesKey
		zo.SharedRules[i] = sr
	}

	for i := range zo.Routes {
		r := zo.Routes[i]
		r.ZoneKey = zo.Zone.ZoneKey
		r.RouteKey = ""
		r.DomainKey = zo.domainKeyMap[r.DomainKey]
		r.SharedRulesKey = zo.sharedRulesKeyMap[r.SharedRulesKey]
		zo.denameifyRules(r.Rules)
		r, err = svc.Route().Create(r)
		if err != nil {
			return nil, err
		}
		zo.Routes[i] = r
	}

	return zo, nil
}
