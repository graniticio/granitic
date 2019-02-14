package ws

import "testing"

func TestCodeMapping(t *testing.T) {
	scd := new(GraniticHTTPStatusCodeDeterminer)

	r := new(Response)
	r.HTTPStatus = 200

	h := scd.DetermineCode(r)

	if h != 200 {
		t.Fail()
	}
}
