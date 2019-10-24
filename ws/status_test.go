package ws

import "testing"

func TestCodeMapping(t *testing.T) {
	scd := NewGraniticHTTPStatusCodeDeterminer()

	r := new(Response)
	r.HTTPStatus = 200

	h := scd.DetermineCode(r)

	if h != 200 {
		t.Fail()
	}
}
