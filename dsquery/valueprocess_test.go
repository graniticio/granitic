package dsquery

import (
	"testing"
)

func TestSQLProcessor(t *testing.T) {

	sp := new(SQLProcessor)

	sp.BoolTrue = "TRUE"
	sp.BoolFalse = "FALSE"

	pvc := ParamValueContext{
		Value: true,
	}

	sp.EscapeParamValue(&pvc)

	if pvc.Value != "TRUE" {
		t.Error()
	}

	pvc = ParamValueContext{
		Value: "abc",
	}

	sp.EscapeParamValue(&pvc)

	if pvc.Value != "'abc'" {
		t.Error()
	}

}
