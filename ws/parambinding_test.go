// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"github.com/graniticio/granitic/v2/types"
	"net/url"
	"testing"
)

func TestQueryArrayBinding(t *testing.T) {
	q := "IA=2,3,5"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)

	tar := struct {
		IA []int64
	}{}

	req := Request{
		QueryParams: qp,
		RequestBody: &tar,
	}

	pb := createParamBinder()
	pb.AutoBindQueryParameters(&req)

	fe := req.FrameworkErrors

	errorCount := len(fe)

	test.ExpectInt(t, errorCount, 0)

	test.ExpectInt(t, len(tar.IA), 3)

}

func TestQueryAutoBinding(t *testing.T) {

	q := "S=s&I=1&I8=8&I16=16&I32=32&I64=64&F32=32.0&F64=64.0&B=true&NS=ns&NI=-64&NF=-10.0E2&NB=false"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	pb.AutoBindQueryParameters(req)

	fe := req.FrameworkErrors

	errorCount := len(fe)

	test.ExpectInt(t, errorCount, 0)

	test.ExpectString(t, bt.S, "s")

	test.ExpectInt(t, bt.I, 1)
	test.ExpectInt(t, int(bt.I8), 8)
	test.ExpectInt(t, int(bt.I16), 16)
	test.ExpectInt(t, int(bt.I32), 32)
	test.ExpectInt(t, int(bt.I64), 64)

	test.ExpectFloat(t, float64(bt.F32), 32.0)
	test.ExpectFloat(t, bt.F64, 64.0)

	test.ExpectBool(t, bt.B, true)

	test.ExpectString(t, bt.NS.String(), "ns")
	test.ExpectBool(t, bt.NB.Bool(), false)
	test.ExpectInt(t, int(bt.NI.Int64()), -64)
	test.ExpectFloat(t, bt.NF.Float64(), -10.0E2)

}

func TestInvalidTargetObjectsAutoBinding(t *testing.T) {

	q := "ITP=0&ITV=1&ITI2"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	it := new(InvalidTargets)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = it

	pb.AutoBindQueryParameters(req)

}

func TestInvalidValuesQueryAutoBinding(t *testing.T) {

	q := "I=A&I8=B&I16=C6&I32=D&I64=F&F32=G&F64=H&B=10&NI=J&NF=K&NB=11"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	pb.AutoBindQueryParameters(req)

	fe := req.FrameworkErrors

	errorCount := len(fe)

	test.ExpectInt(t, errorCount, 11)

}

func TestManualBinding(t *testing.T) {

	q := "SX=s&IX=1&I8X=8&I16X=16&I32X=32&I64X=64&F32X=32.0&F64X=64.0&BX=true&NSX=ns&NIX=-64&NFX=-10.0E2&NBX=false"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	targets := make(map[string]string)

	targets["S"] = "SX"
	targets["I"] = "IX"
	targets["I8"] = "I8X"
	targets["I16"] = "I16X"
	targets["I32"] = "I32X"
	targets["I64"] = "I64X"
	targets["F32"] = "F32X"
	targets["F64"] = "F64X"
	targets["B"] = "BX"
	targets["NS"] = "NSX"
	targets["NI"] = "NIX"
	targets["NF"] = "NFX"
	targets["NB"] = "NBX"

	pb.BindQueryParameters(req, targets)
	fe := req.FrameworkErrors

	errorCount := len(fe)

	test.ExpectInt(t, errorCount, 0)

	test.ExpectString(t, bt.S, "s")

	test.ExpectInt(t, bt.I, 1)
	test.ExpectInt(t, int(bt.I8), 8)
	test.ExpectInt(t, int(bt.I16), 16)
	test.ExpectInt(t, int(bt.I32), 32)
	test.ExpectInt(t, int(bt.I64), 64)

	test.ExpectFloat(t, float64(bt.F32), 32.0)
	test.ExpectFloat(t, bt.F64, 64.0)

	test.ExpectBool(t, bt.B, true)

	test.ExpectString(t, bt.NS.String(), "ns")
	test.ExpectBool(t, bt.NB.Bool(), false)
	test.ExpectInt(t, int(bt.NI.Int64()), -64)
	test.ExpectFloat(t, bt.NF.Float64(), -10.0E2)

}

func TestMissingFieldManualBinding(t *testing.T) {
	q := "S=s"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	targets := make(map[string]string)

	targets["XX"] = "S"

	pb.BindQueryParameters(req, targets)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 1)

}

func TestWrongTypeManualBinding(t *testing.T) {
	q := "S=s"

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	targets := make(map[string]string)

	targets["NB"] = "S"

	pb.BindQueryParameters(req, targets)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 1)

}

func TestSetToNilNillable(t *testing.T) {

	q := "NS=&NI=&NF=&NB="

	v, _ := url.ParseQuery(q)
	qp := NewParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.QueryParams = qp
	req.RequestBody = bt

	pb.AutoBindQueryParameters(req)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 0)

}

func createParamBinder() *ParamBinder {

	fl := new(logging.ConsoleErrorLogger)

	pb := new(ParamBinder)
	pb.FrameworkLogger = fl

	feg := new(FrameworkErrorGenerator)
	feg.FrameworkLogger = fl

	ms := make(map[FrameworkErrorEvent][]string)

	feg.Messages = ms

	pb.FrameworkErrors = feg

	return pb

}

func TestPathBinding(t *testing.T) {

	targets := []string{"S", "I", "I8", "I16", "I32", "I64", "F32", "F64", "B", "NS", "NI", "NF", "NB", "IA", "IS"}
	values := []string{"s", "1", "8", "16", "32", "64", "32.0", "64.0", "true", "ns", "-1", "-64.0", "false", "1,2,3", "A,B,C,D,E"}

	p := NewParamsForPath(targets, values)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.RequestBody = bt

	pb.BindPathParameters(req, p)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 0)

	test.ExpectString(t, bt.S, "s")

	test.ExpectInt(t, bt.I, 1)
	test.ExpectInt(t, int(bt.I8), 8)
	test.ExpectInt(t, int(bt.I16), 16)
	test.ExpectInt(t, int(bt.I32), 32)
	test.ExpectInt(t, int(bt.I64), 64)

	test.ExpectFloat(t, float64(bt.F32), 32.0)
	test.ExpectFloat(t, bt.F64, 64.0)

	test.ExpectBool(t, bt.B, true)

	test.ExpectString(t, bt.NS.String(), "ns")
	test.ExpectBool(t, bt.NB.Bool(), false)
	test.ExpectInt(t, int(bt.NI.Int64()), -1)
	test.ExpectFloat(t, bt.NF.Float64(), -64.0)

	test.ExpectInt(t, len(bt.IA), 3)
	test.ExpectInt(t, len(bt.IS), 5)
}

func TestMorePathTargetsThanValues(t *testing.T) {

	targets := []string{"S", "I", "I8"}
	values := []string{"s", "1"}

	p := NewParamsForPath(targets, values)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.RequestBody = bt

	pb.BindPathParameters(req, p)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 0)

	test.ExpectString(t, bt.S, "s")

	test.ExpectInt(t, bt.I, 1)
	test.ExpectInt(t, int(bt.I8), 0)

}

func TestMorePathValuesThanTargets(t *testing.T) {

	targets := []string{"S", "I"}
	values := []string{"s", "1", "8"}

	p := NewParamsForPath(targets, values)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(Request)
	req.RequestBody = bt

	pb.BindPathParameters(req, p)
	fe := req.FrameworkErrors

	errorCount := len(fe)
	test.ExpectInt(t, errorCount, 0)

	test.ExpectString(t, bt.S, "s")

	test.ExpectInt(t, bt.I, 1)
	test.ExpectInt(t, int(bt.I8), 0)

}

func printErrs(errs []*FrameworkError) {

	for _, e := range errs {
		fmt.Printf("%s->%s %s\n", e.ClientField, e.TargetField, e.Message)
	}
}

type BindingTarget struct {
	S   string
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	F32 float32
	F64 float64
	B   bool
	NS  *types.NilableString
	NI  *types.NilableInt64
	NF  *types.NilableFloat64
	NB  *types.NilableBool
	IA  []int64
	IS  []int64
}

type InvalidTargets struct {
	ITP *InvalidTargetField
	ITV InvalidTargetField
	ITI InvalidInterface
}

type InvalidTargetField struct{}

type InvalidInterface interface{}
