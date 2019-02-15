package httpserver

import (
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccessLogWriter(t *testing.T) {

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
