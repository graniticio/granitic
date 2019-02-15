package httpendpoint

import (
	"bytes"
	"net/http"
	"testing"
)

func TestWritethrough(t *testing.T) {
	rw := new(HTTPResponseWriter)

	rw.rw = new(resWriter)

	rw.WriteHeader(200)

	if !rw.DataSent {
		t.FailNow()
	}

	if rw.Status != 200 {
		t.FailNow()
	}

	rw.Write([]byte{'a'})

	if rw.BytesServed != 1 {
		t.FailNow()
	}

}

type resWriter struct {
	sw bytes.Buffer
}

func (rw *resWriter) Header() http.Header {
	return nil
}

func (rw *resWriter) Write(b []byte) (int, error) {
	return rw.sw.Write(b)
}

func (rw *resWriter) WriteHeader(statusCode int) {

}
