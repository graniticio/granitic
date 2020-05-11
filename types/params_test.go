package types

import (
	"fmt"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestInt16SliceUnparseable(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int16
	}{}

	p := NewSingleValueParams("IS", "1,X,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNotNil(t, err) {

		t.Fatalf("Expected error, got none")
	}

}

func TestNilableStringPointerArray(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []*NilableString
	}{}

	p := NewSingleValueParams("IS", "1,X,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if err != nil {

		t.Fatalf("Expected no error, got %s", err.Error())
	}

	if target.IS[0] == nil {
		t.Fatalf("Expected populated")
	}

}

func TestUnsetSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int64
	}{}

	p := NewSingleValueParams("IS", "")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}

	if target.IS == nil {
		t.Fatalf("Unexpected nil slice %s", err)
	}

	if !test.ExpectInt(t, len(target.IS), 0) {
		t.Fatalf("Unexpected slice length %s", err)
	}

}

func TestUInt16SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []uint16
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestWhitesSpaceSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []uint16
	}{}

	p := NewSingleValueParams("IS", " 1, 2 ,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestInt16SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int16
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestInt8SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int8
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestInt8OverflowSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int8
	}{}

	p := NewSingleValueParams("IS", "1,2,3048")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNotNil(t, err) {
		t.Fatalf("Expected error, got none")
	}

}

func TestRuneSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []rune
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestByteSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []byte
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestInt32SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int32
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestInt64SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []int64
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func TestStringSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []string
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != "1" || a[1] != "2" || a[2] != "3" {
		t.Fatalf("Unexpected value")
	}

}

func TestBoolSlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []bool
	}{}

	p := NewSingleValueParams("IS", "true,false,false")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != true || a[1] != false || a[2] != false {
		t.Fatalf("Unexpected value")
	}

}

func TestUInt64SlicePopulation(t *testing.T) {

	pvi := new(ParamValueInjector)

	target := struct {
		IS []uint64
	}{}

	p := NewSingleValueParams("IS", "1,2,3")

	err := pvi.populateSlice("IS", "IS", p, &target, echoParamError)

	if !test.ExpectNil(t, err) {
		t.Fatalf("Unexpected error %s", err)
	}
	if len(target.IS) != 3 {
		t.Fatalf("Unexpected length")
	}

	a := target.IS

	if a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatalf("Unexpected value")
	}

}

func echoParamError(paramName string, fieldName string, typeName string, params *Params) error {
	return fmt.Errorf(paramName)
}
