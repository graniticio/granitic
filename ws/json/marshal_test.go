package json

import (
	"bytes"
	"context"
	"github.com/graniticio/granitic/v2/types"
	"github.com/graniticio/granitic/v2/ws"
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

func TestMarshalingWriter_MarshalAndWriteFormatOptions(t *testing.T) {

	m := new(MarshalingWriter)

	toMarshal := struct {
		A string
		C int
	}{
		A: "B",
		C: 8,
	}

	var buffer bytes.Buffer

	rw := new(DummyResponseWriter)
	rw.b = &buffer
	m.MarshalAndWrite(toMarshal, rw)

	json := buffer.String()
	expected := "{\"A\":\"B\",\"C\":8}"

	if json != expected {
		t.Errorf("Expected %s, got %s", expected, json)
	}

	m.PrettyPrint = true
	buffer.Reset()

	m.MarshalAndWrite(toMarshal, rw)

	json = buffer.String()
	expected = "{\n\"A\": \"B\",\n\"C\": 8\n}"

	if json != expected {
		t.Errorf("Expected: \n%s\n, got \n%s\n", expected, json)
	}

	m.PrefixString = "!"
	buffer.Reset()

	m.MarshalAndWrite(toMarshal, rw)

	json = buffer.String()
	expected = "{\n!\"A\": \"B\",\n!\"C\": 8\n!}"

	if json != expected {
		t.Errorf("Expected: \n%s\n, got \n%s\n", expected, json)
	}

}

func TestMarshalingWriter_MarshalAndWriteUnsetNilable(t *testing.T) {
	m := new(MarshalingWriter)

	pointerUnset := struct {
		ni *types.NilableInt64
		ns *types.NilableString
		nb *types.NilableBool
		nf *types.NilableFloat64
	}{
		ni: new(types.NilableInt64),
		ns: new(types.NilableString),
		nb: new(types.NilableBool),
		nf: new(types.NilableFloat64),
	}

	var buffer bytes.Buffer

	rw := new(DummyResponseWriter)
	rw.b = &buffer
	err := m.MarshalAndWrite(pointerUnset, rw)

	if err != nil {
		t.Errorf("Could not marshal object with unset nilable %s", err.Error())
	}

	if buffer.String() != "{}" {
		t.Fail()
	}

	valueUnset := struct {
		ni *types.NilableInt64
		ns *types.NilableString
		nb *types.NilableBool
		nf *types.NilableFloat64
	}{}

	buffer.Reset()

	err = m.MarshalAndWrite(valueUnset, rw)

	if err != nil {
		t.Errorf("Could not marshal object with unset nilable %s", err.Error())
	}

	if buffer.String() != "{}" {
		t.Fail()
	}

}

type DummyResponseWriter struct {
	b *bytes.Buffer
}

func (d DummyResponseWriter) Header() http.Header {

	return *new(http.Header)

}

func (d DummyResponseWriter) Write(bytes []byte) (int, error) {

	return d.b.Write(bytes)

}

func (d DummyResponseWriter) WriteHeader(statusCode int) {

}

type target struct {
	A int64
}
