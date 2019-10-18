package httpserver

import (
	"bytes"
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccessLogWriterWithActualFile(t *testing.T) {

	alw := new(AccessLogWriter)

	alw.LogLinePreset = "combined"
	alw.LineBufferSize = 1

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

func TestHTTPRequestLineElements(t *testing.T) {

	alw, fs := logWriterWithBuffer(t, "%U%q %m")

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
