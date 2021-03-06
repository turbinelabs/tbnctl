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

type clusterAdapter struct {
	service.Cluster
}

var _ typelessIface = clusterAdapter{}

func (_ clusterAdapter) Type() objecttype.ObjectType {
	return objecttype.Cluster
}

func (a clusterAdapter) Create(o interface{}) (interface{}, error) {
	return a.Cluster.Create(o.(api.Cluster))
}

func (a clusterAdapter) Get(k string) (interface{}, error) {
	return a.Cluster.Get(api.ClusterKey(k))
}

func (a clusterAdapter) Modify(nxt interface{}) (interface{}, error) {
	return a.Cluster.Modify(nxt.(api.Cluster))
}

func (a clusterAdapter) Delete(k string, cs api.Checksum) error {
	return a.Cluster.Delete(api.ClusterKey(k), cs)
}

func (a clusterAdapter) IndexZeroFilter() interface{} {
	return service.ClusterFilter{}
}

func (a clusterAdapter) Index() ([]interface{}, error) {
	return a.FilteredIndex("", nil)
}

func (a clusterAdapter) FilteredIndex(sliceSep string, attr map[string]string) ([]interface{}, error) {
	f := service.ClusterFilter{}
	populateFilter(&f, attr, sliceSep)

	objs, err := a.Cluster.Index(f)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, 0, len(objs))
	for i := range objs {
		result = append(result, objs[i])
	}

	return result, nil
}

func (a clusterAdapter) Zero() interface{} {
	return api.Cluster{}
}

func (a clusterAdapter) ObjFromString(s string, cd codec.Codec) (interface{}, error) {
	dest := api.Cluster{}
	r := bytes.NewReader([]byte(s))
	err := cd.Decode(r, &dest)
	return dest, err
}

func (a clusterAdapter) Checksum(o interface{}) api.Checksum {
	return o.(api.Cluster).Checksum
}

func mkGetCluster(svc *unifiedSvc) func(k api.ClusterKey) (api.Cluster, error) {
	cache := map[api.ClusterKey]api.Cluster{}
	return func(k api.ClusterKey) (api.Cluster, error) {
		if o, ok := cache[k]; ok {
			return o, nil
		}

		o, err := svc.Cluster().Get(k)
		if err == nil {
			cache[k] = o
		}
		return o, err
	}
}

// underlying allows introspection to the underlying interface for testing
// purposes
func (a clusterAdapter) underlying() interface{} {
	return a.Cluster
}
