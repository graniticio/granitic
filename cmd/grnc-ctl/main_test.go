package main

import (
	"testing"
)

func TestPadRightTo(t *testing.T) {

	s := padRightTo("ABCD", 6)
	

	ex := "ABCD  "

	if s != ex {
		t.Errorf("Expected [%s] was [%s]", ex, s)
	}

}
