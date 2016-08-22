package ws

import (
	"fmt"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws/nillable"
	"net/url"
	"testing"
)

func TestQueryAutoBinding(t *testing.T) {

	q := "S=s&I=1&I8=8&I16=16&I32=32&I64=64&F32=32.0&F64=64.0&B=true&NS=ns&NI=-64&NF=-10.0E2&NB=false"

	v, _ := url.ParseQuery(q)
	qp := NewWsParamsForQuery(v)
	bt := new(BindingTarget)

	pb := createParamBinder()

	req := new(WsRequest)
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

	printErrs(fe)

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

func printErrs(errs []*WsFrameworkError) {

	for _, e := range errs {
		fmt.Printf("%s->%s %s", e.ClientField, e.TargetField, e.Message)
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
	NS  *nillable.NillableString
	NI  *nillable.NillableInt64
	NF  *nillable.NillableFloat64
	NB  *nillable.NillableBool
}
