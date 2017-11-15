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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/turbinelabs/api"
	"github.com/turbinelabs/api/objecttype"
	"github.com/turbinelabs/api/service"
	"github.com/turbinelabs/codec"
	"github.com/turbinelabs/nonstdlib/log/console"
	"github.com/turbinelabs/nonstdlib/must"
	tbnstrings "github.com/turbinelabs/nonstdlib/strings"
	tbntime "github.com/turbinelabs/nonstdlib/time"
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
	DeepDelete(string, api.Checksum, *unifiedSvc) error
	Index() ([]interface{}, error)
	FilteredIndex(string, map[string]string) ([]interface{}, error)
	IndexZeroFilter() interface{}
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

func getAssignmentName(sf reflect.StructField) string {
	t, _ := tbnstrings.Split2(sf.Tag.Get("json"), ",")
	if t == "" {
		return sf.Name
	}

	return t
}

func boolish(s string) bool {
	bval := strings.ToLower(s)
	return (bval == "true" || bval == "t" || bval == "yes" || bval == "y" || bval == "1")
}

func setp(p interface{}, val string) error {
	switch c := p.(type) {
	case **string:
		*c = &val
	case **bool:
		b := boolish(val)
		*c = &b

	case **int8:
		i := int8(must.Int64(strconv.ParseInt(val, 10, 64)))
		*c = &i
	case **int16:
		i := int16(must.Int64(strconv.ParseInt(val, 10, 64)))
		*c = &i
	case **int32:
		i := int32(must.Int64(strconv.ParseInt(val, 10, 64)))
		*c = &i
	case **int64:
		i := must.Int64(strconv.ParseInt(val, 10, 64))
		*c = &i
	case **int:
		i := int(must.Int64(strconv.ParseInt(val, 10, 64)))
		*c = &i

	case **uint8:
		i := uint8(must.Uint64(strconv.ParseUint(val, 10, 64)))
		*c = &i
	case **uint16:
		i := uint16(must.Uint64(strconv.ParseUint(val, 10, 64)))
		*c = &i
	case **uint32:
		i := uint32(must.Uint64(strconv.ParseUint(val, 10, 64)))
		*c = &i
	case **uint64:
		i := must.Uint64(strconv.ParseUint(val, 10, 64))
		*c = &i
	case **uint:
		i := uint(must.Uint64(strconv.ParseUint(val, 10, 64)))
		*c = &i

	case **float32:
		f := float32(must.Float64(strconv.ParseFloat(val, 64)))
		*c = &f
	case **float64:
		f := must.Float64(strconv.ParseFloat(val, 64))
		*c = &f

	case **time.Time:
		ms, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("time must be provided as MS since Unix epoch: %v", err)
		}
		t := tbntime.FromUnixMilli(int64(ms))
		*c = &t

	default:
		return fmt.Errorf("%T is not a supported type.", p)
	}

	return nil
}

func set(fld reflect.Value, val, sliceSep string) error {
	kind := fld.Kind()

	switch kind {
	case reflect.String:
		fld.SetString(val)
	case reflect.Bool:
		fld.SetBool(boolish(val))
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		fld.SetInt(must.Int64(strconv.ParseInt(val, 10, 64)))
	case reflect.Float32, reflect.Float64:
		fld.SetFloat(must.Float64(strconv.ParseFloat(val, 10)))
	case reflect.Struct:
		t := fld.Addr().Interface()
		switch c := t.(type) {
		case *time.Time:
			ms, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("time must be provided as MS since Unix epoch: %v", err)
			}
			*c = tbntime.FromUnixMilli(int64(ms))
			return nil
		}
		return fmt.Errorf("Unable to set %v", kind)
	case reflect.Ptr:
		// Ideally there is a better solution.
		return setp(fld.Addr().Interface(), val)
	case reflect.Slice:
		elements := strings.Split(val, sliceSep)
		slc := reflect.MakeSlice(fld.Type(), len(elements), len(elements))
		fld.Set(slc)

		for i := 0; i < len(elements); i++ {
			err := set(slc.Index(i), elements[i], sliceSep)
			if err != nil {
				return fmt.Errorf("unable to assign element %q in slice", elements[i])
			}
		}

		return nil
	default:
		return fmt.Errorf("Unable to set %v", kind)
	}

	return nil
}

func kindName(v reflect.Value) (string, error) {
	k := v.Kind()
	switch k {
	case reflect.Slice:
		return fmt.Sprintf("slice<%v>", v.Type().Elem().Kind()), nil
	case reflect.Struct:
		i := v.Interface()
		switch i.(type) {
		case time.Time:
			return "time (milliseconds since Unix epoch)", nil
		}
		return "", fmt.Errorf("struct %T is unspported attribute type", i)
	case reflect.Ptr:
		switch v.Interface().(type) {
		case *time.Time:
			return "time (milliseconds since Unix epoch)", nil
		}
		return fmt.Sprintf("%v", v.Type().Elem()), nil
	default:
		return fmt.Sprintf("%v", k), nil
	}
}

func describeFields(f interface{}) map[string]string {
	rv := map[string]string{}

	var (
		v  = reflect.ValueOf(f)
		vt = v.Type() // and this gets the type of the pointer target
	)

	for i := 0; i < vt.NumField(); i++ {
		fsf := vt.Field(i) // filter struct field
		ffv := v.Field(i)  // filter field value

		an := getAssignmentName(fsf)
		n, err := kindName(ffv)
		if err != nil {
			console.Error().Printf("error describing %s: %v", an, err)
		} else {
			rv[an] = n
		}
	}

	return rv
}

// populateFilter takes a pointer to a service.<object>Filter and a map of
// attribute name to value string ad fills in those values. The approach here
// is a bit clumsy and doesn't support complex nested structs. If we introduce
// that into the filter objects we should probably add a go-playground/form
// dependency.
func populateFilter(fptr interface{}, attrs map[string]string, sliceSep string) {
	var (
		v              = reflect.ValueOf(fptr)
		ve             = v.Elem()  // this gets the Value for what our filter point was pointing at
		vet            = ve.Type() // and this gets the type of the pointer target
		filterContents = map[string]reflect.Value{}
	)

	for i := 0; i < vet.NumField(); i++ {
		var (
			fsf = vet.Field(i) // filter struct field
			ffv = ve.Field(i)  // filter field value
		)

		filterContents[getAssignmentName(fsf)] = ffv
	}

	for k, v := range attrs {
		val, ok := filterContents[k]
		if ok {
			err := set(val, v, sliceSep)
			if err != nil {
				console.Error().Fatalf("Unable to set %v: %v\n", k, err)
			}
		}
	}
}
