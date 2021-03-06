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

type sharedRulesAdapter struct {
	service.SharedRules
}

var _ typelessIface = sharedRulesAdapter{}

func (_ sharedRulesAdapter) Type() objecttype.ObjectType {
	return objecttype.SharedRules
}

func (a sharedRulesAdapter) Create(o interface{}) (interface{}, error) {
	return a.SharedRules.Create(o.(api.SharedRules))
}

func (a sharedRulesAdapter) Get(k string) (interface{}, error) {
	return a.SharedRules.Get(api.SharedRulesKey(k))
}

func (a sharedRulesAdapter) Modify(nxt interface{}) (interface{}, error) {
	return a.SharedRules.Modify(nxt.(api.SharedRules))
}

func (a sharedRulesAdapter) Delete(k string, cs api.Checksum) error {
	return a.SharedRules.Delete(api.SharedRulesKey(k), cs)
}

func (a sharedRulesAdapter) IndexZeroFilter() interface{} {
	return service.SharedRulesFilter{}
}

func (a sharedRulesAdapter) Index() ([]interface{}, error) {
	return a.FilteredIndex("", nil)
}

func (a sharedRulesAdapter) FilteredIndex(sliceSep string, attr map[string]string) ([]interface{}, error) {
	f := service.SharedRulesFilter{}
	populateFilter(&f, attr, sliceSep)

	objs, err := a.SharedRules.Index(f)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(objs))
	for i := range objs {
		result = append(result, objs[i])
	}

	return result, nil
}

func (a sharedRulesAdapter) Zero() interface{} {
	return api.SharedRules{}
}

func (a sharedRulesAdapter) ObjFromString(s string, cd codec.Codec) (interface{}, error) {
	dest := api.SharedRules{}
	r := bytes.NewReader([]byte(s))
	err := cd.Decode(r, &dest)
	return dest, err
}

func (a sharedRulesAdapter) Checksum(o interface{}) api.Checksum {
	return o.(api.SharedRules).Checksum
}

func mkGetSharedRules(svc *unifiedSvc) func(k api.SharedRulesKey) (api.SharedRules, error) {
	cache := map[api.SharedRulesKey]api.SharedRules{}
	return func(k api.SharedRulesKey) (api.SharedRules, error) {
		if o, ok := cache[k]; ok {
			return o, nil
		}

		o, err := svc.SharedRules().Get(k)
		if err == nil {
			cache[k] = o
		}
		return o, err
	}
}

// underlying allows introspection to the underlying interface for testing
// purposes
func (a sharedRulesAdapter) underlying() interface{} {
	return a.SharedRules
}
