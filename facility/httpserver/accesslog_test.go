package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/logging"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccessLogWriterWithStdout(t *testing.T) {

	alw := new(AccessLogWriter)

	alw.LogLinePreset = "combined"
	alw.LineBufferSize = 1

	alw.LogPath = stdoutMode

	b := new(UnstructuredLineBuilder)
	b.LogLinePreset = alw.LogLinePreset
	b.LogLineFormat = alw.LogLineFormat

	alw.builder = b

	if err := alw.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	req := new(http.Request)

	end := time.Now()

	start := end.Add(time.Second * -2)

	rw := new(httpendpoint.HTTPResponseWriter)
	rw.DataSent = true
	rw.Status = 200

	alw.LogRequest(context.Background(), req, rw, &start, &end)

	alw.PrepareToStop()

	if ready, _ := alw.ReadyToStop(); ready {
		alw.Stop()
	}
}
func TestAccessLogWriterWithActualFile(t *testing.T) {

	alw := new(AccessLogWriter)

	alw.LogLinePreset = "combined"
	alw.LineBufferSize = 1

	b := new(UnstructuredLineBuilder)
	b.LogLinePreset = alw.LogLinePreset
	b.LogLineFormat = alw.LogLineFormat

	alw.builder = b

	tmp := os.TempDir()

	logPath := filepath.Join(tmp, "test-access.log")

	alw.LogPath = logPath

	if err := alw.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	req := new(http.Request)

	end := time.Now()

	start := end.Add(time.Second * -2)

	rw := new(httpendpoint.HTTPResponseWriter)
	rw.DataSent = true
	rw.Status = 200

	alw.LogRequest(context.Background(), req, rw, &start, &end)

	alw.PrepareToStop()

	if ready, _ := alw.ReadyToStop(); ready {
		alw.Stop()
	}
}

/*
func TestHTTPRequestLineElements(t *testing.T) {

	alw, fs := logWriterWithBuffer(t, "%U%q %m")

	b := new(UnstructuredLineBuilder)
	b.LogLinePreset = alw.LogLinePreset
	b.LogLineFormat = alw.LogLineFormat

	req := new(http.Request)
	req.Method = "POST"
	req.URL, _ = url.Parse("http://localhost:80/test?a=b")

	end := time.Now()

	start := end.Add(time.Second * -2)

	rw := responseWriter(true, 200)

	alw.LogRequest(context.Background(), req, rw, &start, &end)
	alw.PrepareToStop()

	alw.Stop()

	checkContents(t, fs, "/test?a=b POST")
}*/

type ctxKey string

func TestContextValueLogging(t *testing.T) {

	var key ctxKey = "private"

	cxf := new(contextFilter)
	cxf.mappings = make(map[string]ctxKey)

	cxf.mappings["set"] = key

	alw, fs := logWriterWithBuffer(t, "%{set}X %{unset}X")

	alw.ContextFilter = cxf
	alw.builder.SetContextFilter(cxf)

	ctx := context.Background()

	ctx = context.WithValue(ctx, key, "EXPOSED")

	req := new(http.Request)
	end := time.Now()

	start := end.Add(time.Second * -2)

	rw := responseWriter(true, 200)

	alw.LogRequest(ctx, req, rw, &start, &end)
	alw.PrepareToStop()

	alw.Stop()

	checkContents(t, fs, "EXPOSED -")
}

func checkContents(t *testing.T, fs *fileSimulator, ex string) {

	check := ex + "\n"
	actual := fs.buffer.String()

	if actual != check {
		t.Errorf("Unexpected log line. Expected %s Got %s", check, actual)
	}
}

func responseWriter(dataSent bool, status int) *httpendpoint.HTTPResponseWriter {
	rw := new(httpendpoint.HTTPResponseWriter)
	rw.DataSent = dataSent
	rw.Status = status

	return rw
}

func logWriterWithBuffer(t *testing.T, pattern string) (*AccessLogWriter, *fileSimulator) {

	alw := new(AccessLogWriter)
	//alw.LineBufferSize = 1
	alw.LogLineFormat = pattern

	b := new(UnstructuredLineBuilder)
	b.LogLinePreset = alw.LogLinePreset
	b.LogLineFormat = alw.LogLineFormat

	alw.builder = b

	fs := new(fileSimulator)

	alw.openFileFunc = func() (writer closableStringWriter, e error) {

		return fs, nil
	}

	err := alw.StartComponent()

	if err != nil {
		t.Errorf("Unable to create writer for test: %s", err.Error())
	}

	return alw, fs
}

type fileSimulator struct {
	buffer  bytes.Buffer
	Closed  bool
	writing bool
}

func (fs *fileSimulator) WriteString(s string) (n int, err error) {

	fs.writing = true

	n, err = fs.buffer.WriteString(s)

	fs.writing = false

	return n, err
}

func (fs *fileSimulator) Close() error {
	fs.Closed = true

	return nil
}

type contextFilter struct {
	mappings map[string]ctxKey
}

func (cf *contextFilter) Extract(ctx context.Context) logging.FilteredContextData {

	fd := make(logging.FilteredContextData)

	for k, v := range cf.mappings {

		sv := ctx.Value(v)

		fd[k] = fmt.Sprintf("%v", sv)

	}

	return fd
}
