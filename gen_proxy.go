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
	"bytes"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/codec"
)

type proxyAdapter struct {
	service.Proxy
}

var _ typelessIface = proxyAdapter{}

func (_ proxyAdapter) Type() objecttype.ObjectType {
	return objecttype.Proxy
}

func (a proxyAdapter) Create(o interface{}) (interface{}, error) {
	return a.Proxy.Create(o.(api.Proxy))
}

func (a proxyAdapter) Get(k string) (interface{}, error) {
	return a.Proxy.Get(api.ProxyKey(k))
}

func (a proxyAdapter) Modify(nxt interface{}) (interface{}, error) {
	return a.Proxy.Modify(nxt.(api.Proxy))
}

func (a proxyAdapter) Delete(k string, cs api.Checksum) error {
	return a.Proxy.Delete(api.ProxyKey(k), cs)
}

func (a proxyAdapter) IndexZeroFilter() interface{} {
	return service.ProxyFilter{}
}

func (a proxyAdapter) Index() ([]interface{}, error) {
	return a.FilteredIndex("", nil)
}

func (a proxyAdapter) FilteredIndex(sliceSep string, attr map[string]string) ([]interface{}, error) {
	f := service.ProxyFilter{}
	populateFilter(&f, attr, sliceSep)

	objs, err := a.Proxy.Index(f)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(objs))
	for i := range objs {
		result = append(result, objs[i])
	}

	return result, nil
}

func (a proxyAdapter) Zero() interface{} {
	return api.Proxy{}
}

func (a proxyAdapter) ObjFromString(s string, cd codec.Codec) (interface{}, error) {
	dest := api.Proxy{}
	r := bytes.NewReader([]byte(s))
	err := cd.Decode(r, &dest)
	return dest, err
}

func (a proxyAdapter) Checksum(o interface{}) api.Checksum {
	return o.(api.Proxy).Checksum
}

func mkGetProxy(svc *unifiedSvc) func(k api.ProxyKey) (api.Proxy, error) {
	cache := map[api.ProxyKey]api.Proxy{}
	return func(k api.ProxyKey) (api.Proxy, error) {
		if o, ok := cache[k]; ok {
			return o, nil
		}

		o, err := svc.Proxy().Get(k)
		if err == nil {
			cache[k] = o
		}
		return o, err
	}
}

// underlying allows introspection to the underlying interface for testing
// purposes
func (a proxyAdapter) underlying() interface{} {
	return a.Proxy
}
