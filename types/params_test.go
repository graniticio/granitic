package types

import (
	"fmt"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

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

func TestSlicePopulation(t *testing.T) {

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

func echoParamError(paramName string, fieldName string, typeName string, params *Params) error {
	return fmt.Errorf(paramName)
}
