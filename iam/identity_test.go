package iam

import "testing"

func TestNewAuthenticatedIdentity(t *testing.T) {

	var a ClientIdentity
	a = NewAuthenticatedIdentity("id")

	if !a.Authenticated() {
		t.FailNow()
	}
}
