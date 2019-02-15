package main

import (
	"testing"
)

func TestPadRightTo(t *testing.T) {

	s := padRightTo("ABCD", 2)
	ex := "  ABCD"

	if s != ex {
		t.Errorf("Expected [%s] was [%s]", ex, s)
	}

}
