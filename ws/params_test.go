package ws

import (
	"net/url"
	"testing"
)

func TestDetectValues(t *testing.T) {

	q := "a=b"

	v, _ := url.ParseQuery(q)
	qp := NewWsParamsForQuery(v)

	if !qp.Exists("a") {
		t.Errorf("Expected key 'a' to be present")
	}

	if qp.Exists("b") {
		t.Errorf("Did not expect key 'b' to be present")
	}

}

func TestDetectMultiple(t *testing.T) {

	q := "a=b&a=c&x=y"

	v, _ := url.ParseQuery(q)
	qp := NewWsParamsForQuery(v)

	if !qp.Exists("a") {
		t.Errorf("Expected key 'a' to be present")
	}

	if !qp.MultipleValues("a") {
		t.Errorf("Expected 'a' to have multiple values")
	}

	if qp.MultipleValues("x") {
		t.Errorf("Expected 'x' to have single value")
	}

}

func TestStringValues(t *testing.T) {

	q := "a=b&a=c&x=y"

	v, _ := url.ParseQuery(q)
	qp := NewWsParamsForQuery(v)

	a, err := qp.StringValue("a")

	if err != nil {
		t.Errorf("Did not expect an error when trying to key value for 'a'")
	}

	if a != "c" {
		t.Errorf("Expected 'a' to have value b, instead was %s", a)
	}

	x, err := qp.StringValue("x")

	if err != nil {
		t.Errorf("Did not expect an error when trying to key value for 'x'")
	}

	if x != "y" {
		t.Errorf("Expected 'x' to have value y, instead was %s", a)
	}

	_, err = qp.StringValue("z")

	if err == nil {
		t.Errorf("Expected an error when looking for missing key 'z'")
	}

}
