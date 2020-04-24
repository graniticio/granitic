// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/instance"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

const percent = "%"
const hyphen = "-"
const unsupportedPlaceholder = "???"
const presetCommonName = "common"
const presetCommonFormat = "%h %l %u %t \"%r\" %s %b"
const presetCombinedName = "combined"

// PresetCombinedFormat is the log format used when AccessLogWriter.LogLinePreset is set to combined. Similar to the Apache HTTP preset of the same name.
const PresetCombinedFormat = "%h %l %u %t \"%r\" %s %b \"%{Referer}i\" \"%{User-agent}i\""

const presetFrameworkName = "framework"

// PresetFrameworkFormat is the log format used when AccessLogWriter.LogLinePreset is set to framework. Uses the X-Forwarded-For header to show all
// IP addresses that the request has been proxied for (useful for services that sit behind multiple load-balancers and proxies) and logs
// processing time in microseconds.
const PresetFrameworkFormat = "%h XFF[%{X-Forwarded-For}i] %l %u [%{02/Jan/2006:15:04:05 Z0700}t] \"%m %U%q\" %s %bB %{us}Tμs"

const formatRegex = "\\%[a-zA-Z]|\\%\\%|\\%{[^}]*}[a-zA-Z]"
const varModifiedRegex = "\\%{([^}]*)}([a-zA-Z])"
const commonLogDateFormat = "[02/Jan/2006:15:04:05 -0700]"

const stdoutMode = "STDOUT"

// AccessLogWriter is a component able to asynchronously write an Apache HTTPD style access log. See the top of this GoDoc page for more information.
type AccessLogWriter struct {
	logFile closableStringWriter

	openFileFunc func() (closableStringWriter, error)

	// The path of the log file to be written to (and created if required)
	LogPath string

	// The format of each log line. See the top of this GoDoc page for supported formats. Mutually exclusive with LogLinePreset.
	LogLineFormat string

	// A pre-defined format. Supported values are framework or combined. Mutually exclusive with LogLineFormat.
	LogLinePreset string

	//The number of lines that can be buffered for asynchronous writing to the log file before calls to LogRequest block.
	//Setting to zero or less makes calls to LogRequest synchronous
	LineBufferSize int

	// Whether or not timestamps should be converted to UTC before they are written to the access log.
	UtcTimes bool

	// A component able to extract information from a context.Context into a loggable format
	ContextFilter logging.ContextFilter

	builder LineBuilder

	lines chan string
	state ioc.ComponentState
}

// LogRequest generates an access log line according the configured format. As long as the number of log lines waiting to
// be written to the file does not exceed the value of AccessLogWriter.LineBufferSize, this method will return immediately.
func (alw *AccessLogWriter) LogRequest(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) {

	if alw.state != ioc.RunningState {
		return
	}

	alw.lines <- alw.builder.BuildLine(ctx, req, res, rec, fin)

}

// RegisterInstanceID receives  the instance ID of the current application and passes it down to the LineBuilder
func (alw *AccessLogWriter) RegisterInstanceID(i *instance.Identifier) {
	if alw.builder != nil {
		alw.builder.SetInstanceID(i)
	}
}

// LineBuilder is a component able to generate an access log entry ready to be written to a file or stream
type LineBuilder interface {
	BuildLine(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) string
	SetContextFilter(cf logging.ContextFilter)
	Init() error
	SetInstanceID(i *instance.Identifier)
}

// StartComponent parses the specified log format, sets up a channel to buffer lines for asynchrnous writing and opens the log file. An error
// is returned if any of these steps fails.
func (alw *AccessLogWriter) StartComponent() error {

	if alw.openFileFunc == nil {
		alw.openFileFunc = alw.openFile
	}

	if alw.state != ioc.StoppedState {
		return nil
	}

	alw.state = ioc.StartingState

	if alw.LineBufferSize > 0 {
		alw.lines = make(chan string, alw.LineBufferSize)
	} else {
		alw.lines = make(chan string)
	}

	err := alw.builder.Init()
	alw.builder.SetContextFilter(alw.ContextFilter)

	if err != nil {
		return err
	}

	alw.logFile, err = alw.openFileFunc()

	if err != nil {
		return err
	}

	go alw.watchLineBuffer()

	alw.state = ioc.RunningState

	return err

}

func (alw *AccessLogWriter) watchLineBuffer() {
	for {
		line := <-alw.lines

		f := alw.logFile

		if f != nil {
			f.WriteString(line)
		}

	}
}

func (alw *AccessLogWriter) openFile() (closableStringWriter, error) {
	logPath := alw.LogPath

	if len(strings.TrimSpace(logPath)) == 0 {
		return nil, errors.New("HTTP server access log is enabled, but no path to a log file specified")
	}

	if logPath == stdoutMode {
		return uncloseable{os.Stdout}, nil
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func intMax(x, y int) int {
	if x > y {
		return x
	}

	return y

}

// PrepareToStop settings state to 'Stopping'
func (alw *AccessLogWriter) PrepareToStop() {
	alw.state = ioc.StoppingState

}

// ReadyToStop returns true if the log line buffer is empty
func (alw *AccessLogWriter) ReadyToStop() (bool, error) {

	l := len(alw.lines)

	if l == 0 {
		return true, nil
	}

	return false, fmt.Errorf("%s waiting to writing %d lines to the log file", accessLogWriterName, l)

}

// Stop closes the log file and message channel
func (alw *AccessLogWriter) Stop() error {

	if alw.lines != nil {
		close(alw.lines)
	}

	alw.state = ioc.StoppedState

	if alw.logFile != nil {
		return alw.logFile.Close()
	}

	return nil
}

type closableStringWriter interface {
	WriteString(s string) (n int, err error)
	Close() error
}

type uncloseable struct {
	c closableStringWriter
}

func (u uncloseable) WriteString(s string) (n int, err error) {
	return u.c.WriteString(s)
}

func (u uncloseable) Close() error {
	return nil
}
