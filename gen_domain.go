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

type domainAdapter struct {
	service.Domain
}

var _ typelessIface = domainAdapter{}

func (_ domainAdapter) Type() objecttype.ObjectType {
	return objecttype.Domain
}

func (a domainAdapter) Create(o interface{}) (interface{}, error) {
	return a.Domain.Create(o.(api.Domain))
}

func (a domainAdapter) Get(k string) (interface{}, error) {
	return a.Domain.Get(api.DomainKey(k))
}

func (a domainAdapter) Modify(nxt interface{}) (interface{}, error) {
	return a.Domain.Modify(nxt.(api.Domain))
}

func (a domainAdapter) Delete(k string, cs api.Checksum) error {
	return a.Domain.Delete(api.DomainKey(k), cs)
}

func (a domainAdapter) IndexZeroFilter() interface{} {
	return service.DomainFilter{}
}

func (a domainAdapter) Index() ([]interface{}, error) {
	return a.FilteredIndex("", nil)
}

func (a domainAdapter) FilteredIndex(sliceSep string, attr map[string]string) ([]interface{}, error) {
	f := service.DomainFilter{}
	populateFilter(&f, attr, sliceSep)

	objs, err := a.Domain.Index(f)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(objs))
	for i := range objs {
		result = append(result, objs[i])
	}

	return result, nil
}

func (a domainAdapter) Zero() interface{} {
	return api.Domain{}
}

func (a domainAdapter) ObjFromString(s string, cd codec.Codec) (interface{}, error) {
	dest := api.Domain{}
	r := bytes.NewReader([]byte(s))
	err := cd.Decode(r, &dest)
	return dest, err
}

func (a domainAdapter) Checksum(o interface{}) api.Checksum {
	return o.(api.Domain).Checksum
}

func mkGetDomain(svc *unifiedSvc) func(k api.DomainKey) (api.Domain, error) {
	cache := map[api.DomainKey]api.Domain{}
	return func(k api.DomainKey) (api.Domain, error) {
		if o, ok := cache[k]; ok {
			return o, nil
		}

		o, err := svc.Domain().Get(k)
		if err == nil {
			cache[k] = o
		}
		return o, err
	}
}

// underlying allows introspection to the underlying interface for testing
// purposes
func (a domainAdapter) underlying() interface{} {
	return a.Domain
}
