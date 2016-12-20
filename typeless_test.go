package main

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
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
