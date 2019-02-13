package main

import (
	"fmt"
	"testing"
)

func TestPadRightTo(t *testing.T) {

	s := padRightTo("ABCD", 2)
	ex := "  ABCD"

	if s != ex {
		fmt.Errorf("Expected [%s] was [%s]", ex, s)
	}

}
