package xml

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/graniticio/granitic/v3/ws"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestMarshallAndWrite(t *testing.T) {

	mw := new(MarshalingWriter)
	rw := new(resWriter)

	mw.MarshalAndWrite(target{A: 1}, rw)

	if rw.sw.String() != "<content><a>1</a></content>" {
		t.Fail()
	}

}

func TestUnmarshalling(t *testing.T) {

	r := new(http.Request)
	sr := strings.NewReader("<content><a>1</a></content>")

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
	XMLName xml.Name `xml:"content"`
	A       int64    `xml:"a"`
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
