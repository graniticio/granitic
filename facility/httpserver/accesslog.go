// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

type logFormatPlaceHolder int

const (
	unsupported = iota
	remoteHost
	ctxValue
	clientID
	userID
	receivedTime
	requestLine
	statusCode
	bytesReturned
	bytesReturnedClf
	requestHeader
	percentSymbol
	method
	path
	query
	processTimeMicro
	processTime
)

type logLineTokenType int

const (
	textToken = iota
	placeholderToken
	placeholderWithVar
)

type logLineToken struct {
	tokenType       logLineTokenType
	placeholderType logFormatPlaceHolder
	content         string
	variable        string
}

func newTextLogLineElement(text string) *logLineToken {

	e := new(logLineToken)
	e.tokenType = textToken
	e.content = text

	return e
}

func newPlaceholderLineElement(phType logFormatPlaceHolder) *logLineToken {

	e := new(logLineToken)
	e.tokenType = placeholderToken
	e.placeholderType = phType

	return e
}

func newPlaceholderWithVarLineElement(phType logFormatPlaceHolder, variable string) *logLineToken {

	e := new(logLineToken)
	e.tokenType = placeholderWithVar
	e.placeholderType = phType
	e.variable = variable

	return e
}

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

	elements []*logLineToken
	lines    chan string
	state    ioc.ComponentState
}

// LogRequest generates an access log line according the configured format. As long as the number of log lines waiting to
// be written to the file does not exceed the value of AccessLogWriter.LineBufferSize, this method will return immediately.
func (alw *AccessLogWriter) LogRequest(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) {

	if alw.state != ioc.RunningState {
		return
	}

	alw.lines <- alw.buildLine(ctx, req, res, rec, fin)

}

func (alw *AccessLogWriter) buildLine(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) string {
	var b bytes.Buffer

	var cv logging.FilteredContextData

	if alw.ContextFilter != nil {
		// Extract loggable information from the context
		cv = alw.ContextFilter.Extract(ctx)
	}

	if alw.UtcTimes {
		utcRec := rec.UTC()
		utcFin := fin.UTC()

		rec = &utcRec
		fin = &utcFin
	}

	for _, e := range alw.elements {

		switch e.tokenType {
		case textToken:
			b.WriteString(e.content)
		case placeholderToken:
			b.WriteString(alw.findValue(ctx, e, req, res, rec, fin))
		case placeholderWithVar:
			b.WriteString(alw.findValueWithVar(ctx, cv, e, req, res, rec, fin))
		}
	}

	b.WriteString("\n")

	return b.String()

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

	err := alw.configureLogFormat()

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

	if logPath == "stdout" {
		return os.Stdout, nil
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (alw *AccessLogWriter) configureLogFormat() error {

	f := alw.LogLineFormat
	pre := alw.LogLinePreset

	if f == "" && pre == "" {
		return errors.New("you must specify either a format for access log lines or the name of a preset format (neither has been provided)")
	}

	if f != "" {
		//Custom log mode - ignore the preset
		pre = ""
	}

	if pre != "" {

		if pre == presetCommonName {
			return alw.parseFormat(presetCommonFormat)

		} else if pre == presetCombinedName {
			return alw.parseFormat(PresetCombinedFormat)

		} else if pre == presetFrameworkName {
			return alw.parseFormat(PresetFrameworkFormat)

		} else {
			message := fmt.Sprintf("%s is not a supported preset for access log lines", pre)
			return errors.New(message)
		}

	}

	return alw.parseFormat(f)
}

func (alw *AccessLogWriter) parseFormat(format string) error {

	lineRe := regexp.MustCompile(formatRegex)
	varRe := regexp.MustCompile(varModifiedRegex)

	placeholders := lineRe.FindAllString(format, -1)
	textFragments := lineRe.Split(format, -1)
	firstMatch := lineRe.FindStringIndex(format)
	var startsWithPh bool

	if len(firstMatch) > 0 {
		startsWithPh = (firstMatch[0] == 0) && textFragments[0] != ""
	}

	phCount := len(placeholders)
	tCount := len(textFragments)

	maxCount := intMax(phCount, tCount)

	for i := 0; i < maxCount; i++ {

		phAvail := i < phCount
		tAvail := i < tCount
		var err error

		if phAvail && tAvail {

			ph := placeholders[i]
			text := textFragments[i]

			if startsWithPh {
				err = alw.addPlaceholder(ph, varRe)
				alw.addTextElement(text)
			} else {
				alw.addTextElement(text)
				err = alw.addPlaceholder(ph, varRe)
			}

		} else if phAvail {
			ph := placeholders[i]
			err = alw.addPlaceholder(ph, varRe)

		} else if tAvail {
			text := textFragments[i]
			alw.addTextElement(text)
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func (alw *AccessLogWriter) addTextElement(text string) {

	if text != "" {
		e := newTextLogLineElement(text)
		alw.elements = append(alw.elements, e)
	}
}

func (alw *AccessLogWriter) addPlaceholder(ph string, re *regexp.Regexp) error {

	if len(ph) == 2 {

		formatTypeCode := ph[1:2]

		lfph := alw.mapPlaceholder(formatTypeCode)

		if lfph == unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		}

		e := newPlaceholderLineElement(lfph)
		alw.elements = append(alw.elements, e)

	} else {

		v := re.FindStringSubmatch(ph)
		arg := v[1]
		formatTypeCode := v[2]

		lfph := alw.mapPlaceholder(formatTypeCode)

		if lfph == unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		}

		e := newPlaceholderWithVarLineElement(lfph, arg)
		alw.elements = append(alw.elements, e)

	}

	return nil
}

func intMax(x, y int) int {
	if x > y {
		return x
	}

	return y

}

func (alw *AccessLogWriter) mapPlaceholder(ph string) logFormatPlaceHolder {

	switch ph {
	default:
		return unsupported
	case "%":
		return percentSymbol
	case "b":
		return bytesReturnedClf
	case "B":
		return bytesReturned
	case "D":
		return processTimeMicro
	case "h":
		return remoteHost
	case "i":
		return requestHeader
	case "l":
		return clientID
	case "m":
		return method
	case "q":
		return query
	case "r":
		return requestLine
	case "s":
		return statusCode
	case "t":
		return receivedTime
	case "T":
		return processTime
	case "u":
		return userID
	case "U":
		return path
	case "X":
		return ctxValue
	}

}

func (alw *AccessLogWriter) findValueWithVar(ctx context.Context, cd logging.FilteredContextData, element *logLineToken, req *http.Request, res *httpendpoint.HTTPResponseWriter, received *time.Time, finished *time.Time) string {
	switch element.placeholderType {
	case requestHeader:
		return alw.requestHeader(element.variable, req)

	case receivedTime:
		return received.Format(element.variable)

	case processTime:

		switch element.variable {
		case "s":
			return alw.processTime(received, finished, time.Second)
		case "us":
			return alw.processTime(received, finished, time.Microsecond)
		case "ms":
			return alw.processTime(received, finished, time.Millisecond)
		default:
			return "??"

		}
	case ctxValue:
		return alw.ctxValue(cd, element.variable)

	default:
		return unsupportedPlaceholder

	}
}

func (alw *AccessLogWriter) findValue(ctx context.Context, element *logLineToken, req *http.Request, res *httpendpoint.HTTPResponseWriter, received *time.Time, finished *time.Time) string {

	switch element.placeholderType {

	case percentSymbol:
		return percent

	case bytesReturnedClf:
		if res.BytesServed == 0 {
			return hyphen
		}

		return (strconv.Itoa(res.BytesServed))

	case bytesReturned:
		return (strconv.Itoa(res.BytesServed))

	case remoteHost:
		return req.RemoteAddr

	case clientID:
		return hyphen

	case userID:
		return alw.userID(ctx)

	case method:
		return req.Method

	case path:
		return req.URL.Path

	case query:
		return alw.query(req)

	case requestLine:
		return alw.requestLine(req)

	case receivedTime:
		return received.Format(commonLogDateFormat)

	case statusCode:
		return strconv.Itoa(res.Status)

	case processTimeMicro:
		return alw.processTime(received, finished, time.Microsecond)

	case processTime:
		return alw.processTime(received, finished, time.Second)

	default:
		return unsupportedPlaceholder

	}

}

func (alw *AccessLogWriter) ctxValue(cd logging.FilteredContextData, key string) string {

	if cd == nil || cd[key] == "" {
		return hyphen
	}

	return cd[key]

}

func (alw *AccessLogWriter) processTime(rec *time.Time, fin *time.Time, unit time.Duration) string {
	spent := fin.Sub(*rec)

	return strconv.FormatInt(int64(spent/unit), 10)
}

func (alw *AccessLogWriter) query(req *http.Request) string {

	q := req.URL.RawQuery

	if q == "" {
		return q
	}

	return "?" + q

}

func (alw *AccessLogWriter) requestHeader(name string, req *http.Request) string {

	value := req.Header.Get(name)

	if value == "" {
		return hyphen

	}

	return value
}

func (alw *AccessLogWriter) requestLine(req *http.Request) string {
	return fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)
}

func (alw *AccessLogWriter) userID(ctx context.Context) string {
	return hyphen
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

	if alw.logFile != nil {
		return alw.logFile.Close()
	}

	if alw.lines != nil {
		close(alw.lines)
	}

	alw.state = ioc.StoppedState

	return nil
}

type closableStringWriter interface {
	WriteString(s string) (n int, err error)
	Close() error
}
