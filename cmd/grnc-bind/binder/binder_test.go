package binder

import (
	"fmt"
	"testing"
)

func TestBinderContains(t *testing.T) {

	b := new(Binder)

	if !b.contains([]string{"A", "B", "C"}, "B") {
		t.Fail()
	}

	if b.contains([]string{"A", "B", "C"}, "D") {
		t.Fail()
	}
}

func TestRemoveEscapes(t *testing.T) {

	b := new(Binder)

	s := b.removeEscapes("++T").(string)

	if s != "+T" {

		fmt.Println(s)
		t.Error()
	}

	s = b.removeEscapes("??T").(string)

	if s != "?T" {
		t.Error()
	}

}

func TestCheckRefDetection(t *testing.T) {

	b := new(Binder)

	if b.isRef(true) {
		t.Error()
	}

	if !b.isRef("ref:myThing") {
		t.Error()
	}

	if !b.isRef("r:myThing") {
		t.Error()
	}

	if !b.isRef("+MyThing") {
		t.Error()
	}

	if b.isRef("++MyThing") {
		t.Error()
	}

}

func TestCheckConfDetection(t *testing.T) {

	b := new(Binder)

	if b.isPromise(true) {
		t.Error()
	}

	if !b.isPromise("conf:myThing") {
		t.Error()
	}

	if !b.isPromise("c:myThing") {
		t.Error()
	}

	if !b.isPromise("?MyThing") {
		t.Error()
	}

	if b.isPromise("??MyThing") {
		t.Error()
	}

}
