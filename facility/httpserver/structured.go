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
	Prefix string
	Fields []*AccessLogJsonField
	Suffix string
	UTC    bool
}

// A AccessLogJsonField defines the rules for outputting a single field in a JSON-formatted application log entry
type AccessLogJsonField struct {
	Name      string
	Content   string
	Arg       string
	generator AccessLogValueGenerator
}

const (
	ctxVal = "CONTEXT_VALUE"
)

// ValidateJSONFields checks that the configuration of a JSON application log entry is correct
func ValidateJSONFields(fields []*AccessLogJsonField) error {

	allowed := types.NewOrderedStringSet([]string{ctxVal})

	for _, f := range fields {

		if !allowed.Contains(f.Content) {
			return fmt.Errorf("%s is not a supported content type for a JSON log field. Allowed values are %v", f.Content, allowed.Contents())
		}

		if f.Content == ctxVal && strings.TrimSpace(f.Arg) == "" {
			return fmt.Errorf("you must specify an Arg when using JSON fields with the content type %s (the key of the value to be extracted from the context filter)", ctxVal)
		}

	}

	return nil

}

// CreateMapBuilder builds a component able to generate a log entry based on the rules in the supplied fields.
func CreateMapBuilder(cfg *AccessLogJSONConfig) (*AccessLogMapBuilder, error) {

	mb := new(AccessLogMapBuilder)

	mb.cfg = cfg

	for _, f := range cfg.Fields {

		switch f.Content {
		case ctxVal:
			mb.RequiresContextFilter = true
			f.generator = mb.ctxValGenerator
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
	}

	for _, f := range mb.cfg.Fields {

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
}

// ValueGenerator functions are able to generate a value for a field in a JSON formatted log entry
type AccessLogValueGenerator func(lineContext *LineContext, field *AccessLogJsonField) interface{}

func (mb *AccessLogMapBuilder) ctxValGenerator(lineContext *LineContext, field *AccessLogJsonField) interface{} {

	if lineContext.FilteredContext != nil {
		return lineContext.FilteredContext[field.Arg]
	}

	return ""
}
