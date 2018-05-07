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
	"testing"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/test/assert"
)

type initZoneParseRouteTestCase struct {
	in     string
	errStr string
	route  route
}

func TestParseDomainsCollapses(t *testing.T) {
	got, err := parseDomains([]string{
		"foo:80=bar",
		"foo:80=baz",
		"bar:443=blegga:blagga",
	})
	want := []domain{
		{name: hostPort{"foo", 80}, aliases: api.DomainAliases{"bar", "baz"}},
		{name: hostPort{"bar", 443}, aliases: api.DomainAliases{"blegga", "blagga"}},
	}
	assert.Nil(t, err)
	assert.DeepEqual(t, got, want)
}

func TestParseDomainsBadHostPort(t *testing.T) {
	got, err := parseDomains([]string{"bogus=bar"})
	assert.Nil(t, got)
	assert.ErrorContains(t, err, `malformed domain/port "bogus" in domains argument "bogus=bar"`)
}

func TestParseProxiesCollapses(t *testing.T) {
	got, err := parseProxies([]string{
		"foo=foo1:8081",
		"foo=foo2:8082",
		"bar=bar1:80",
	})
	want := []proxy{
		{name: "foo", domains: []hostPort{{"foo1", 8081}, {"foo2", 8082}}},
		{name: "bar", domains: []hostPort{{"bar1", 80}}},
	}
	assert.Nil(t, err)
	assert.DeepEqual(t, got, want)
}

func TestValidateProxiesOK(t *testing.T) {
	err := validateProxies(
		[]proxy{{name: "foo", domains: []hostPort{{"bar", 80}}}},
		[]route{{domain: hostPort{"bar", 80}}},
	)
	assert.Nil(t, err)
}

func TestValidateProxiesUnknown(t *testing.T) {
	err := validateProxies(
		[]proxy{{name: "foo", domains: []hostPort{{"bar", 80}}}},
		nil,
	)
	assert.ErrorContains(t, err, "proxy foo refers to unknown domain bar:80")
}

func TestValidateRoutesOK(t *testing.T) {
	err := validateRoutes(
		[]route{
			{domain: hostPort{"bar", 80}, path: "/"},
			{domain: hostPort{"bar", 80}, path: "/baz"},
			{domain: hostPort{"bar", 443}, path: "/baz"},
		},
	)
	assert.Nil(t, err)
}

func TestValidateRoutesDuplicate(t *testing.T) {
	err := validateRoutes(
		[]route{
			{domain: hostPort{"bar", 80}, path: "/baz"},
			{domain: hostPort{"bar", 80}, path: "/baz"},
		},
	)
	assert.ErrorContains(t, err, "route bar:80/baz declared more than once")
}

func TestParseRoute(t *testing.T) {
	for _, tc := range []initZoneParseRouteTestCase{
		{
			in:     "",
			errStr: `malformed route ""`,
		},
		{
			in:     "aksjfghdslkfjashflsdkfjhdsf",
			errStr: `malformed route "aksjfghdslkfjashflsdkfjhdsf"`,
		},
		{
			in:     "example.com:80",
			errStr: `malformed route "example.com:80"`,
		},
		{
			in:     "example.com:80=",
			errStr: `malformed route "example.com:80="`,
		},
		{
			in:     "example.com:80=:nocluster",
			errStr: `empty cluster name in route argument "example.com:80=:nocluster"`,
		},
		{
			in:     "example.com=exampleService",
			errStr: `malformed domain/port "example.com" in route argument "example.com=exampleService"`,
		},
		{
			in:     "example.com:blegga=exampleService",
			errStr: `malformed domain/port "example.com:blegga" in route argument "example.com:blegga=exampleService"`,
		},
		{
			in:     "api.example.com:443/api/users=userService:stage=prod:",
			errStr: `malformed metadata "" in route argument "api.example.com:443/api/users=userService:stage=prod:"`,
		},
		{
			in:     "api.example.com:443/api/users=userService:",
			errStr: `malformed metadata "" in route argument "api.example.com:443/api/users=userService:"`,
		},
		{
			in:     "api.example.com:443/api/users=userService:=missingkey",
			errStr: `malformed metadata "=missingkey" in route argument "api.example.com:443/api/users=userService:=missingkey"`,
		},
		{
			in: "example.com:80=exampleService",
			route: route{
				domain:   hostPort{"example.com", 80},
				path:     "/",
				cluster:  "exampleService",
				metadata: api.Metadata{},
			},
		},
		{
			in: "example.com:80=exampleService:foo=bar",
			route: route{
				domain:  hostPort{"example.com", 80},
				path:    "/",
				cluster: "exampleService",
				metadata: api.Metadata{
					api.Metadatum{"foo", "bar"},
				},
			},
		},
		{
			in: "api.example.com:443/api=apiService",
			route: route{
				domain:   hostPort{"api.example.com", 443},
				path:     "/api",
				cluster:  "apiService",
				metadata: api.Metadata{},
			},
		},
		{
			in: "api.example.com:443/api/users=userService:stage=prod:version=1.0",
			route: route{
				domain:  hostPort{"api.example.com", 443},
				path:    "/api/users",
				cluster: "userService",
				metadata: api.Metadata{
					api.Metadatum{"stage", "prod"},
					api.Metadatum{"version", "1.0"},
				},
			},
		},
		{
			in: "api.example.com:443/api/users=userService:stage=prod:version=",
			route: route{
				domain:  hostPort{"api.example.com", 443},
				path:    "/api/users",
				cluster: "userService",
				metadata: api.Metadata{
					api.Metadatum{"stage", "prod"},
					api.Metadatum{"version", ""},
				},
			},
		},
		{
			in: "api.example.com:443/api/users=userService:stage=prod:version",
			route: route{
				domain:  hostPort{"api.example.com", 443},
				path:    "/api/users",
				cluster: "userService",
				metadata: api.Metadata{
					api.Metadatum{"stage", "prod"},
					api.Metadatum{"version", ""},
				},
			},
		},
		{
			in: "api.example.com:443/api/users=userService:stage=prod:version=something with =:foo=bar",
			route: route{
				domain:  hostPort{"api.example.com", 443},
				path:    "/api/users",
				cluster: "userService",
				metadata: api.Metadata{
					api.Metadatum{"stage", "prod"},
					api.Metadatum{"version", "something with ="},
					api.Metadatum{"foo", "bar"},
				},
			},
		},
	} {
		assert.Group("TestParseRoute: "+tc.in, t, func(g *assert.G) {
			route, err := parseRoute(tc.in)
			if tc.errStr == "" {
				assert.Nil(g, err)
			} else {
				assert.ErrorContains(g, err, tc.errStr)
			}
			assert.DeepEqual(g, route, tc.route)
		})
	}
}
