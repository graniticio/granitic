package json

import (
	"context"
	"github.com/graniticio/granitic/v3/ws"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestErrorMarshalling(t *testing.T) {

	e := new(ws.ServiceErrors)
	e.AddError(ws.NewCategorisedError(ws.Client, "AAA", "error"))

	gef := new(GraniticJSONErrorFormatter)

	f := gef.FormatErrors(e).(map[string]interface{})

	g := f["General"]

	if g == nil {
		t.Fail()
	}

}

func TestUnmarshalling(t *testing.T) {

	r := new(http.Request)
	sr := strings.NewReader("{\"A\": 1}")

	rc := ioutil.NopCloser(sr)

	r.Body = rc

	um := new(Unmarshaller)

	wsr := new(ws.Request)
	wsr.RequestBody = new(target)

	um.Unmarshall(context.Background(), r, wsr)

	tar := wsr.RequestBody.(*target)

	if tar.A != 1 {
		t.Fail()
	}

}

type target struct {
	A int64
}
