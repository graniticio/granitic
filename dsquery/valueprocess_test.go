package dsquery

import (
	"testing"
)

func TestSQLProcessor(t *testing.T) {

	sp := new(SQLProcessor)

	sp.BoolTrue = "TRUE"
	sp.BoolFalse = "FALSE"

	pvc := paramValueContext{
		Value: true,
	}

	sp.EscapeParamValue(&pvc)

	if pvc.Value != "TRUE" {
		t.Error()
	}

	pvc = paramValueContext{
		Value: "abc",
	}

	sp.EscapeParamValue(&pvc)

	if pvc.Value != "'abc'" {
		t.Error()
	}

}
