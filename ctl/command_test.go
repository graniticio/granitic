package ctl

import (
	"github.com/graniticio/granitic/v2/ws"
	"testing"
)

func TestErrorGenerators(t *testing.T) {

	m := "message"

	e := NewCommandClientError(m)

	if e.Message != m {
		t.Error()
	}

	if e.Category != ws.Client {
		t.Error()
	}

	e = NewCommandUnexpectedError(m)

	if e.Message != m {
		t.Error()
	}

	if e.Category != ws.Unexpected {
		t.Error()
	}

	e = NewCommandLogicError(m)

	if e.Message != m {
		t.Error()
	}

	if e.Category != ws.Logic {
		t.Error()
	}

}
