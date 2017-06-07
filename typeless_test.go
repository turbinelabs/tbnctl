/*
Copyright 2017 Turbine Labs, Inc.

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
	"time"

	"github.com/golang/mock/gomock"

	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/nonstdlib/ptr"
	tbntime "github.com/turbinelabs/nonstdlib/time"
	"github.com/turbinelabs/test/assert"
)

var (
	minObjectTypeID = 1
	maxObjectTypeID = 1
)

func init() {
	for {
		_, err := objecttype.FromID(maxObjectTypeID)
		if err != nil {
			break
		}
		maxObjectTypeID++
	}
}

func TestNewTypelessIface(t *testing.T) {
	for i := minObjectTypeID; i < maxObjectTypeID; i++ {
		ctrl := gomock.NewController(assert.Tracing(t))
		all := service.NewMockAll(ctrl)
		admin := service.NewMockAdmin(ctrl)
		fin := ctrl.Finish

		svc := &unifiedSvc{all, admin}
		ot, _ := objecttype.FromID(i)

		if ot == objecttype.Org {
			iface := newTypelessIface(svc, ot)
			assert.Nil(t, iface)
			fin()
			continue
		}

		if ot == objecttype.User {
			mu := service.NewMockUser(ctrl)
			admin.EXPECT().User().Return(mu)
			iface := newTypelessIface(svc, ot)
			assert.SameInstance(t, iface.(testadapter).underlying(), mu)
			fin()
			continue
		}

		var want interface{}

		switch ot {
		case objecttype.Zone:
			m := service.NewMockZone(ctrl)
			all.EXPECT().Zone().Return(m)
			want = m
		case objecttype.Proxy:
			m := service.NewMockProxy(ctrl)
			all.EXPECT().Proxy().Return(m)
			want = m
		case objecttype.Domain:
			m := service.NewMockDomain(ctrl)
			all.EXPECT().Domain().Return(m)
			want = m
		case objecttype.SharedRules:
			m := service.NewMockSharedRules(ctrl)
			all.EXPECT().SharedRules().Return(m)
			want = m
		case objecttype.Route:
			m := service.NewMockRoute(ctrl)
			all.EXPECT().Route().Return(m)
			want = m
		case objecttype.Cluster:
			m := service.NewMockCluster(ctrl)
			all.EXPECT().Cluster().Return(m)
			want = m
		}

		got := newTypelessIface(svc, ot)
		assert.NonNil(t, want)
		assert.SameInstance(t, got.(testadapter).underlying(), want)

		fin()
	}
}

type testadapter interface {
	underlying() interface{}
}

func TestDescribeFields(t *testing.T) {
	type test struct {
		I    int
		Is   []int
		S    string
		Ss   []string
		T    time.Time
		Tp   *time.Time
		I64  int64
		I64p *int64
	}

	got := describeFields(test{})
	assert.DeepEqual(t, got, map[string]string{
		"I":    "int",
		"Is":   "slice<int>",
		"S":    "string",
		"Ss":   "slice<string>",
		"T":    "time (milliseconds since Unix epoch)",
		"Tp":   "time (milliseconds since Unix epoch)",
		"I64":  "int64",
		"I64p": "int64",
	})
}

func TestPopulateFilter(t *testing.T) {
	type dest struct {
		I     int
		S     string
		F     float32
		B     bool
		Is    []int
		Ss    []string
		Bs    []bool
		Ip    *int
		I32ps []*int32
		Sp    *string
		Fp    *float64
		Bp    *bool
		T     time.Time
		Tp    *time.Time
	}

	attrs := map[string]string{
		"I":     "1234",
		"S":     "some string",
		"F":     "3e10",
		"B":     "y",
		"Is":    "1,2,3,4",
		"Ss":    "one,two,three,four",
		"Bs":    "yes,no,0,true,false",
		"Ip":    "1234",
		"I32ps": "1,2,3,4",
		"Sp":    "test string",
		"Fp":    "40.0",
		"Bp":    "false",
		"T":     "10000",
		"Tp":    "1000",
	}

	got := dest{}
	populateFilter(&got, attrs, ",")

	want := dest{
		1234,
		"some string",
		30000000000,
		true,
		[]int{1, 2, 3, 4},
		[]string{"one", "two", "three", "four"},
		[]bool{true, false, false, true, false},
		ptr.Int(1234),
		[]*int32{ptr.Int32(1), ptr.Int32(2), ptr.Int32(3), ptr.Int32(4)},
		ptr.String("test string"),
		ptr.Float64(40),
		ptr.Bool(false),
		tbntime.FromUnixMilli(10000),
		ptr.Time(tbntime.FromUnixMilli(1000)),
	}

	assert.DeepEqual(t, got, want)
}
