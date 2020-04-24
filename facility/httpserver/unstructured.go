// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

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

// UnstructuredLineBuilder creates strings that correspond to traditional line-based entries in an accesss log, where the
// meaning is inferred by position of each field in the line
type UnstructuredLineBuilder struct {

	// The format of each log line. See the top of this GoDoc page for supported formats. Mutually exclusive with LogLinePreset.
	LogLineFormat string

	// A pre-defined format. Supported values are framework or combined. Mutually exclusive with LogLineFormat.
	LogLinePreset string

	cf         logging.ContextFilter
	elements   []*logLineToken
	utcTimes   bool
	instanceID *instance.Identifier
}

// SetInstanceID records the ID of the current instance
func (ulb *UnstructuredLineBuilder) SetInstanceID(i *instance.Identifier) {
	ulb.instanceID = i
}

// SetContextFilter provides access to a component that has limited access to data stored in a context.Context
func (ulb *UnstructuredLineBuilder) SetContextFilter(cf logging.ContextFilter) {
	ulb.cf = cf
}

// BuildLine creates a text representation of an access log entry ready to log to a file or stream
func (ulb *UnstructuredLineBuilder) BuildLine(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) string {
	var b bytes.Buffer

	var cv logging.FilteredContextData

	if ulb.cf != nil {
		// Extract loggable information from the context
		cv = ulb.cf.Extract(ctx)
	}

	if ulb.utcTimes {
		utcRec := rec.UTC()
		utcFin := fin.UTC()

		rec = &utcRec
		fin = &utcFin
	}

	for _, e := range ulb.elements {

		switch e.tokenType {
		case textToken:
			b.WriteString(e.content)
		case placeholderToken:
			b.WriteString(ulb.findValue(ctx, e, req, res, rec, fin))
		case placeholderWithVar:
			b.WriteString(ulb.findValueWithVar(ctx, cv, e, req, res, rec, fin))
		}
	}

	b.WriteString("\n")

	return b.String()

}

func (ulb *UnstructuredLineBuilder) parseFormat(format string) error {

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
				err = ulb.addPlaceholder(ph, varRe)
				ulb.addTextElement(text)
			} else {
				ulb.addTextElement(text)
				err = ulb.addPlaceholder(ph, varRe)
			}

		} else if phAvail {
			ph := placeholders[i]
			err = ulb.addPlaceholder(ph, varRe)

		} else if tAvail {
			text := textFragments[i]
			ulb.addTextElement(text)
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func (ulb *UnstructuredLineBuilder) addTextElement(text string) {

	if text != "" {
		e := newTextLogLineElement(text)
		ulb.elements = append(ulb.elements, e)
	}
}

func (ulb *UnstructuredLineBuilder) addPlaceholder(ph string, re *regexp.Regexp) error {

	if len(ph) == 2 {

		formatTypeCode := ph[1:2]

		lfph := ulb.mapPlaceholder(formatTypeCode)

		if lfph == unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		}

		e := newPlaceholderLineElement(lfph)
		ulb.elements = append(ulb.elements, e)

	} else {

		v := re.FindStringSubmatch(ph)
		arg := v[1]
		formatTypeCode := v[2]

		lfph := ulb.mapPlaceholder(formatTypeCode)

		if lfph == unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		}

		e := newPlaceholderWithVarLineElement(lfph, arg)
		ulb.elements = append(ulb.elements, e)

	}

	return nil
}

// Init checks that a valid format for access log lines has been provided
func (ulb *UnstructuredLineBuilder) Init() error {

	f := ulb.LogLineFormat
	pre := ulb.LogLinePreset

	if f == "" && pre == "" {
		return errors.New("you must specify either a format for access log lines or the name of a preset format (neither has been provided)")
	}

	if f != "" {
		//Custom log mode - ignore the preset
		pre = ""
	}

	if pre != "" {

		if pre == presetCommonName {
			return ulb.parseFormat(presetCommonFormat)

		} else if pre == presetCombinedName {
			return ulb.parseFormat(PresetCombinedFormat)

		} else if pre == presetFrameworkName {
			return ulb.parseFormat(PresetFrameworkFormat)

		} else {
			message := fmt.Sprintf("%s is not a supported preset for access log lines", pre)
			return errors.New(message)
		}

	}

	return ulb.parseFormat(f)
}

func (ulb *UnstructuredLineBuilder) mapPlaceholder(ph string) logFormatPlaceHolder {

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

func (ulb *UnstructuredLineBuilder) findValueWithVar(ctx context.Context, cd logging.FilteredContextData, element *logLineToken, req *http.Request, res *httpendpoint.HTTPResponseWriter, received *time.Time, finished *time.Time) string {
	switch element.placeholderType {
	case requestHeader:
		return ulb.requestHeader(element.variable, req)

	case receivedTime:
		return received.Format(element.variable)

	case processTime:

		switch element.variable {
		case "s":
			return ulb.processTime(received, finished, time.Second)
		case "us":
			return ulb.processTime(received, finished, time.Microsecond)
		case "ms":
			return ulb.processTime(received, finished, time.Millisecond)
		default:
			return "??"

		}
	case ctxValue:
		return ulb.ctxValue(cd, element.variable)

	default:
		return unsupportedPlaceholder

	}
}

func (ulb *UnstructuredLineBuilder) findValue(ctx context.Context, element *logLineToken, req *http.Request, res *httpendpoint.HTTPResponseWriter, received *time.Time, finished *time.Time) string {

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
		return ulb.userID(ctx)

	case method:
		return req.Method

	case path:
		return req.URL.Path

	case query:
		return ulb.query(req)

	case requestLine:
		return ulb.requestLine(req)

	case receivedTime:
		return received.Format(commonLogDateFormat)

	case statusCode:
		return strconv.Itoa(res.Status)

	case processTimeMicro:
		return ulb.processTime(received, finished, time.Microsecond)

	case processTime:
		return ulb.processTime(received, finished, time.Second)

	default:
		return unsupportedPlaceholder

	}

}

func (ulb *UnstructuredLineBuilder) ctxValue(cd logging.FilteredContextData, key string) string {

	if cd == nil || cd[key] == "" {
		return hyphen
	}

	return cd[key]

}

func (ulb *UnstructuredLineBuilder) processTime(rec *time.Time, fin *time.Time, unit time.Duration) string {
	spent := fin.Sub(*rec)

	return strconv.FormatInt(int64(spent/unit), 10)
}

func (ulb *UnstructuredLineBuilder) query(req *http.Request) string {

	q := req.URL.RawQuery

	if q == "" {
		return q
	}

	return "?" + q

}

func (ulb *UnstructuredLineBuilder) requestHeader(name string, req *http.Request) string {

	value := req.Header.Get(name)

	if value == "" {
		return hyphen

	}

	return value
}

func (ulb *UnstructuredLineBuilder) requestLine(req *http.Request) string {
	return fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)
}

func (ulb *UnstructuredLineBuilder) userID(ctx context.Context) string {
	return hyphen
}
