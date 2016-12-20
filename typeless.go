package main

import (
	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/codec"
)

type unifiedSvc struct {
	service.All
	service.Admin
}

//go:generate genny -in adapter.genny -out gen_user.go -pkg $GOPACKAGE gen "__type__=user __Type__=User"
//go:generate genny -in adapter.genny -out gen_zone.go -pkg $GOPACKAGE gen "__type__=zone __Type__=Zone"
//go:generate genny -in adapter.genny -out gen_proxy.go -pkg $GOPACKAGE gen "__type__=proxy __Type__=Proxy"
//go:generate genny -in adapter.genny -out gen_domain.go -pkg $GOPACKAGE gen "__type__=domain __Type__=Domain"
//go:generate genny -in adapter.genny -out gen_route.go -pkg $GOPACKAGE gen "__type__=route __Type__=Route"
//go:generate genny -in adapter.genny -out gen_sharedrules.go -pkg $GOPACKAGE gen "__type__=sharedrules __Type__=SharedRules"
//go:generate genny -in adapter.genny -out gen_cluster.go -pkg $GOPACKAGE gen "__type__=cluster __Type__=Cluster"

type typelessIface interface {
	Type() objecttype.ObjectType

	ObjFromString(string, codec.Codec) (interface{}, error)
	Checksum(interface{}) api.Checksum
	Zero() interface{}

	Create(interface{}) (interface{}, error)
	Get(string) (interface{}, error)
	Modify(interface{}) (interface{}, error)
	Delete(string, api.Checksum) error
	Index() ([]interface{}, error)
}

func newTypelessIface(svc *unifiedSvc, ot objecttype.ObjectType) typelessIface {
	switch ot {
	case objecttype.Zone:
		return zoneAdapter{svc.All.Zone()}
	case objecttype.Proxy:
		return proxyAdapter{svc.All.Proxy()}
	case objecttype.Domain:
		return domainAdapter{svc.All.Domain()}
	case objecttype.SharedRules:
		return sharedrulesAdapter{svc.All.SharedRules()}
	case objecttype.Route:
		return routeAdapter{svc.All.Route()}
	case objecttype.Cluster:
		return clusterAdapter{svc.All.Cluster()}
	case objecttype.User:
		return userAdapter{svc.Admin.User()}
	}

	return nil
}
