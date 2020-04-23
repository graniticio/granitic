// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// A JSONLineBuilder is a component able to take a message to be written to a log file and format it as JSON document
type JSONLineBuilder struct {
	Config     *AccessLogJSONConfig
	MapBuilder *AccessLogMapBuilder
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (jlb *JSONLineBuilder) BuildLine(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) string {

	m := jlb.MapBuilder.BuildLine(ctx, req, res, rec, fin)
	cfg := jlb.Config

	entry, _ := json.Marshal(m)

	return cfg.Prefix + string(entry) + cfg.Suffix
}

// StartComponent checks that a context filter has been injected (if the field configuration needs on)
func (jlb *JSONLineBuilder) Init() error {

	mb := jlb.MapBuilder

	if mb.RequiresContextFilter && mb.contextFilter == nil {
		return fmt.Errorf("your JSON application logging configuration includes fields that display information from the contecxt, but no component is available that implements logging.ContextFilter")
	}

	return nil
}

//SetContextFilter provides the formatter with access selected data from a context
func (jlb *JSONLineBuilder) SetContextFilter(cf logging.ContextFilter) {
	jlb.MapBuilder.contextFilter = cf
}

// AccessLogJSONConfig defines the fields to be included in a  JSON-formatted application log entry
type AccessLogJSONConfig struct {
	Prefix       string
	Fields       [][]string
	ParsedFields []*AccessLogJsonField
	Suffix       string
	UTC          bool
}

// A AccessLogJsonField defines the rules for outputting a single field in a JSON-formatted application log entry
type AccessLogJsonField struct {
	Name      string
	Content   string
	Arg       string
	generator AccessLogValueGenerator
}

const (
	ctxVal         = "CONTEXT_VALUE"
	remote         = "REMOTE"
	reqHeader      = "REQ_HEADER"
	received       = "RECEIVED"
	httpMethod     = "HTTP_METHOD"
	reqPath        = "PATH"
	queryString    = "QUERY"
	status         = "STATUS"
	bytesOut       = "BYTES_OUT"
	processingTime = "PROCESS_TIME"
	reqLine        = "REQUEST_LINE"
)

const (
	seconds = "SECONDS"
	milli   = "MILLI"
	micro   = "MICRO"
)

// ValidateJSONFields checks that the configuration of a JSON application log entry is correct
func ValidateJSONFields(fields []*AccessLogJsonField) error {

	allowed := types.NewOrderedStringSet([]string{ctxVal, remote, reqHeader, received, httpMethod, reqPath, queryString, status, bytesOut, processingTime, reqLine})

	argNeeded := types.NewOrderedStringSet([]string{ctxVal, reqHeader, received, processingTime})

	for _, f := range fields {

		if !allowed.Contains(f.Content) {
			return fmt.Errorf("%s is not a supported content type for a JSON log field. Allowed values are %v", f.Content, allowed.Contents())
		}

		if argNeeded.Contains(f.Content) && strings.TrimSpace(f.Arg) == "" {

			return fmt.Errorf("you must specify an Arg when using JSON fields with the content type %s", f.Content)

		}

		if f.Content == processingTime && f.Arg != seconds && f.Arg != milli && f.Arg != micro {

			return fmt.Errorf("the arg for fields of type %s must be one of %s %s %s ", f.Content, seconds, milli, micro)

		}

	}

	return nil

}

//ConvertFields converts from the config representation of a field list to the internal version
func ConvertFields(unparsed [][]string) []*AccessLogJsonField {

	l := len(unparsed)

	if l == 0 {
		return make([]*AccessLogJsonField, 0)
	}

	allParsed := make([]*AccessLogJsonField, l)

	for i, raw := range unparsed {

		parsed := new(AccessLogJsonField)
		fcount := len(raw)

		if fcount > 0 {
			parsed.Name = raw[0]
		}

		if fcount > 1 {
			parsed.Content = raw[1]
		}

		if fcount > 2 {
			parsed.Arg = raw[2]
		}

		allParsed[i] = parsed

	}

	return allParsed
}

// CreateMapBuilder builds a component able to generate a log entry based on the rules in the supplied fields.
func CreateMapBuilder(cfg *AccessLogJSONConfig) (*AccessLogMapBuilder, error) {

	mb := new(AccessLogMapBuilder)

	mb.cfg = cfg

	for _, f := range cfg.ParsedFields {

		switch f.Content {
		case ctxVal:
			mb.RequiresContextFilter = true
			f.generator = mb.ctxValGenerator
		case remote:
			f.generator = mb.remoteGenerator
		case reqHeader:
			f.generator = mb.reqHeaderGenerator
		case received:
			f.generator = mb.receivedTimeGenerator
		case httpMethod:
			f.generator = mb.methodGenerator
		case reqPath:
			f.generator = mb.pathGenerator
		case queryString:
			f.generator = mb.queryGenerator
		case status:
			f.generator = mb.statusGenerator
		case bytesOut:
			f.generator = mb.bytesOutGenerator
		case processingTime:
			if f.Arg == seconds {
				f.generator = mb.processSecondsGenerator
			} else if f.Arg == milli {
				f.generator = mb.processMilliGenerator
			} else {
				f.generator = mb.processMicroGenerator
			}
		case reqLine:
			f.generator = mb.reqLineGenerator
		}
	}

	return mb, nil
}

// AccessLogMapBuilder creates a map[string]interface{} representing a log entry, ready for JSON encoding
type AccessLogMapBuilder struct {
	cfg                   *AccessLogJSONConfig
	contextFilter         logging.ContextFilter
	RequiresContextFilter bool
}

// Build creates a map and populates it
func (mb *AccessLogMapBuilder) BuildLine(ctx context.Context, req *http.Request, res *httpendpoint.HTTPResponseWriter, rec *time.Time, fin *time.Time) map[string]interface{} {

	var fcd logging.FilteredContextData

	outer := make(map[string]interface{})

	if mb.RequiresContextFilter && mb.contextFilter != nil && ctx != nil {
		fcd = mb.contextFilter.Extract(ctx)
	}

	c := LineContext{
		FilteredContext: fcd,
		Request:         req,
		ResponseWriter:  res,
		Received:        rec,
		Finished:        fin,
		Ctx:             &ctx,
	}

	for _, f := range mb.cfg.ParsedFields {

		outer[f.Name] = f.generator(&c, f)

	}

	return outer

}

type LineContext struct {
	FilteredContext logging.FilteredContextData
	Request         *http.Request
	ResponseWriter  *httpendpoint.HTTPResponseWriter
	Received        *time.Time
	Finished        *time.Time
	Ctx             *context.Context
}

// ValueGenerator functions are able to generate a value for a field in a JSON formatted log entry
type AccessLogValueGenerator func(lineContext *LineContext, field *AccessLogJsonField) interface{}

func (mb *AccessLogMapBuilder) ctxValGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {

	if lineContext.FilteredContext != nil {
		return lineContext.FilteredContext[field.Arg]
	}

	return ""
}

func (mb *AccessLogMapBuilder) remoteGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Request.RemoteAddr
}

func (mb *AccessLogMapBuilder) reqHeaderGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Request.Header.Get(field.Arg)
}

func (mb *AccessLogMapBuilder) receivedTimeGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Received.Format(field.Arg)
}

func (mb *AccessLogMapBuilder) methodGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Request.Method
}

func (mb *AccessLogMapBuilder) pathGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Request.URL.Path
}

func (mb *AccessLogMapBuilder) queryGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return lineContext.Request.URL.RawQuery
}

func (mb *AccessLogMapBuilder) statusGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return strconv.Itoa(lineContext.ResponseWriter.Status)
}

func (mb *AccessLogMapBuilder) bytesOutGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return strconv.Itoa(lineContext.ResponseWriter.BytesServed)
}

func (mb *AccessLogMapBuilder) processSecondsGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return processTimeGen(lineContext.Received, lineContext.Finished, time.Second)
}

func (mb *AccessLogMapBuilder) processMilliGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return processTimeGen(lineContext.Received, lineContext.Finished, time.Millisecond)
}

func (mb *AccessLogMapBuilder) processMicroGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {
	return processTimeGen(lineContext.Received, lineContext.Finished, time.Microsecond)
}

func processTimeGen(rec *time.Time, fin *time.Time, unit time.Duration) string {
	spent := fin.Sub(*rec)

	return strconv.FormatInt(int64(spent/unit), 10)
}

func (mb *AccessLogMapBuilder) reqLineGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {

	req := lineContext.Request

	return fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)
}
