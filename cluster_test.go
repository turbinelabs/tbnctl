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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/test/assert"
)

var (
	ex = errors.New("snatoesanhans")
	cl = api.Cluster{Name: "wheeeee"}
)

func mkAdapter(t *testing.T) (clusterAdapter, *service.MockCluster, func()) {
	ctrl := gomock.NewController(assert.Tracing(t))
	m := service.NewMockCluster(ctrl)
	return clusterAdapter{m}, m, ctrl.Finish
}

func TestType(t *testing.T) {
	a, _, fin := mkAdapter(t)
	defer fin()

	assert.Equal(t, a.Type(), objecttype.Cluster)
}

func TestUnderlying(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	assert.SameInstance(t, a.underlying(), m)
}

func TestCreate(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	m.EXPECT().Create(cl).Return(cl, ex)
	got, gotErr := a.Create(cl)

	assert.True(t, got.(api.Cluster).Equals(cl))
	assert.DeepEqual(t, gotErr, ex)
}

func TestGet(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	k := "asonetuh"
	m.EXPECT().Get(api.ClusterKey(k)).Return(cl, ex)
	got, gotErr := a.Get(k)
	assert.True(t, got.(api.Cluster).Equals(cl))
	assert.Equal(t, gotErr, ex)
}

func TestModify(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	n := cl
	n.Checksum = api.Checksum{"saoentuhasonteuh"}
	m.EXPECT().Modify(n).Return(cl, ex)

	got, gotErr := a.Modify(n)
	assert.True(t, got.(api.Cluster).Equals(cl))
	assert.Equal(t, gotErr, ex)
}

func TestDelete(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	k := "asonteuha"
	cs := api.Checksum{"snthsnthsnth"}

	m.EXPECT().Delete(api.ClusterKey(k), cs).Return(ex)
	gotErr := a.Delete(k, cs)
	assert.Equal(t, gotErr, ex)
}

func TestIndex(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	c1 := api.Cluster{Name: "c1"}
	c2 := api.Cluster{Name: "c2"}

	m.EXPECT().Index(service.ClusterFilter{}).Return([]api.Cluster{c1, c2}, nil)
	got, gotErr := a.Index()

	assert.Nil(t, gotErr)
	assert.DeepEqual(t, got, []interface{}{c1, c2})
}

func TestIndexErr(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	m.EXPECT().Index(service.ClusterFilter{}).Return([]api.Cluster{{}, {}, {}}, ex)
	got, gotErr := a.Index()

	assert.Nil(t, got)
	assert.Equal(t, gotErr, ex)
}

func TestFilteredIndex(t *testing.T) {
	a, m, fin := mkAdapter(t)
	defer fin()

	m.EXPECT().Index(service.ClusterFilter{Name: "bob"}).Return([]api.Cluster{{}, {}, {}}, ex)
	got, gotErr := a.FilteredIndex(",", map[string]string{"name": "bob"})

	assert.Nil(t, got)
	assert.Equal(t, gotErr, ex)
}

func TestZero(t *testing.T) {
	a, _, fin := mkAdapter(t)
	defer fin()

	z := a.Zero()
	assert.True(t, z.(api.Cluster).Equals(api.Cluster{}))
}

func TestChecksum(t *testing.T) {
	a, _, fin := mkAdapter(t)
	defer fin()

	c := api.Cluster{Checksum: api.Checksum{"asonetuhaosnethuasothe"}}
	assert.Equal(t, a.Checksum(c), c.Checksum)
}
