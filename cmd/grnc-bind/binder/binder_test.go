package binder

import (
	"fmt"
	"github.com/graniticio/granitic/v2/test"
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

	s = b.removeEscapes("$$T").(string)

	if s != "$T" {
		t.Error()
	}

}

func TestDefaultValueExtraction(t *testing.T) {

	b := new(Binder)
	b.compileRegexes()

	s, v := b.extractDefaultValue("a.b.c.d")

	if s != "a.b.c.d" || v != "" {
		t.Fail()
	}

	s, v = b.extractDefaultValue("a.b.c.d(true)")

	if s != "a.b.c.d" {
		t.Fail()
	}

	if v != "true" {
		t.Fail()
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

	if !b.isPromise("$MyThing") {
		t.Error()
	}

	if b.isPromise("$$MyThing") {
		t.Error()
	}

}

func TestManifestValidation(t *testing.T) {

	m := new(Manifest)

	if !test.ExpectNotNil(t, validateManifest(m)) {

		t.Errorf("Expected error on missing namespace")

	}

	m.Namespace = " A "

	if !test.ExpectNotNil(t, validateManifest(m)) {

		t.Errorf("Expected error on string padded namespace")

	}

	m.Namespace = "A#B"

	if !test.ExpectNotNil(t, validateManifest(m)) {

		t.Errorf("Expected error on illegal characters in namespace")

	}

	m.Namespace = "1AB"

	if !test.ExpectNotNil(t, validateManifest(m)) {

		t.Errorf("Expected error on namespace starting with number")

	}

	m.Namespace = "grnc"

	if !test.ExpectNil(t, validateManifest(m)) {

		t.Errorf("Expected no error with valid namespace")

	}

	def := new(definition)

	m.ExternalFacilities = make(map[string]*definition)

	m.ExternalFacilities["2Invalid!"] = def

	if !test.ExpectNotNil(t, validateManifest(m)) {

		t.Errorf("Expected error on invalid facility name")

	}

}
