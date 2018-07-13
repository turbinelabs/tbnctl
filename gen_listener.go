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

type listenerAdapter struct {
	service.Listener
}

var _ typelessIface = listenerAdapter{}

func (_ listenerAdapter) Type() objecttype.ObjectType {
	return objecttype.Listener
}

func (a listenerAdapter) Create(o interface{}) (interface{}, error) {
	return a.Listener.Create(o.(api.Listener))
}

func (a listenerAdapter) Get(k string) (interface{}, error) {
	return a.Listener.Get(api.ListenerKey(k))
}

func (a listenerAdapter) Modify(nxt interface{}) (interface{}, error) {
	return a.Listener.Modify(nxt.(api.Listener))
}

func (a listenerAdapter) Delete(k string, cs api.Checksum) error {
	return a.Listener.Delete(api.ListenerKey(k), cs)
}

func (a listenerAdapter) IndexZeroFilter() interface{} {
	return service.ListenerFilter{}
}

func (a listenerAdapter) Index() ([]interface{}, error) {
	return a.FilteredIndex("", nil)
}

func (a listenerAdapter) FilteredIndex(sliceSep string, attr map[string]string) ([]interface{}, error) {
	f := service.ListenerFilter{}
	populateFilter(&f, attr, sliceSep)

	objs, err := a.Listener.Index(f)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(objs))
	for i := range objs {
		result = append(result, objs[i])
	}

	return result, nil
}

func (a listenerAdapter) Zero() interface{} {
	return api.Listener{}
}

func (a listenerAdapter) ObjFromString(s string, cd codec.Codec) (interface{}, error) {
	dest := api.Listener{}
	r := bytes.NewReader([]byte(s))
	err := cd.Decode(r, &dest)
	return dest, err
}

func (a listenerAdapter) Checksum(o interface{}) api.Checksum {
	return o.(api.Listener).Checksum
}

func mkGetListener(svc *unifiedSvc) func(k api.ListenerKey) (api.Listener, error) {
	cache := map[api.ListenerKey]api.Listener{}
	return func(k api.ListenerKey) (api.Listener, error) {
		if o, ok := cache[k]; ok {
			return o, nil
		}

		o, err := svc.Listener().Get(k)
		if err == nil {
			cache[k] = o
		}
		return o, err
	}
}

// underlying allows introspection to the underlying interface for testing
// purposes
func (a listenerAdapter) underlying() interface{} {
	return a.Listener
}
