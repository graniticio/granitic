// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graniticio/granitic/v2/types"
	"strings"
	"time"
)

// A JSONLogFormatter is a component able to take a message to be written to a log file and format it as JSON document
type JSONLogFormatter struct {
	// A component able to extract information from a context.Context into a loggable format
	ContextFilter ContextFilter
	Config        *JSONConfig
	MapBuilder    *MapBuilder
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (jlf *JSONLogFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {

	m := jlf.MapBuilder.Build(ctx, levelLabel, loggerName, message)
	cfg := jlf.Config

	entry, _ := json.Marshal(m)

	return cfg.Prefix + string(entry) + cfg.Suffix
}

//SetContextFilter provides the formatter with access selected data from a context
func (jlf *JSONLogFormatter) SetContextFilter(cf ContextFilter) {
	jlf.ContextFilter = cf
}

// JSONConfig defines the fields to be included in a  JSON-formatted application log entry
type JSONConfig struct {
	Prefix string
	Fields []*JSONField
	Suffix string
	UTC    bool
}

// A JSONField defines the rules for outputting a single field in a JSON-formatted application log entry
type JSONField struct {
	Name      string
	Content   string
	Arg       string
	generator ValueGenerator
}

const (
	message   = "MESSAGE"
	comp      = "COMPONENT_NAME"
	timestamp = "TIMESTAMP"
	level     = "LEVEL"
	ctxVal    = "CONTEXT_VALUE"
)

// ValidateJSONFields checks that the configuration of a JSON application log entry is correct
func ValidateJSONFields(fields []*JSONField) error {

	allowed := types.NewOrderedStringSet([]string{message, comp, timestamp, level, ctxVal})

	for i, f := range fields {

		if strings.TrimSpace(f.Name) == "" {
			return fmt.Errorf("JSON log field at position %d in the array of fields is missing a Name", i)
		}

		if !allowed.Contains(f.Content) {
			return fmt.Errorf("%s is not a supported content type for a JSON log field. Allowed values are %v", f.Content, allowed.Contents())
		}

		if f.Content == ctxVal && strings.TrimSpace(f.Arg) == "" {
			return fmt.Errorf("you must specify an Arg when using JSON fields with the content type %s (the key of the value to be extracted from the context filter)", ctxVal)
		}

		if f.Content == timestamp {

			if strings.TrimSpace(f.Arg) == "" {
				return fmt.Errorf("you must specify an Arg when using JSON fields with the content type %s (a standard Go date/time layout string)", timestamp)
			}

			ft := time.Now().Format(f.Arg)

			if _, err := time.Parse(f.Arg, ft); err != nil {
				return fmt.Errorf("unable to use layout [%s] as a timestamp layout", f.Arg)
			}

		}

	}

	return nil

}

// CreateMapBuilder builds a component able to generate a log entry based on the rules in the supplied fields.
func CreateMapBuilder(cfg *JSONConfig) (*MapBuilder, error) {

	mb := new(MapBuilder)

	mb.cfg = cfg

	for _, f := range cfg.Fields {

		switch f.Content {
		case message:
			f.generator = messageGenerator
		case comp:
			f.generator = componentGenerator
		case level:
			f.generator = levelGenerator
		case timestamp:
			if cfg.UTC {
				f.generator = utcTimestampGenerator
			} else {
				f.generator = localTimestampGenerator
			}

		}

	}

	return mb, nil
}

// MapBuilder creates a map[string]interface{} representing a log entry, ready for JSON encoding
type MapBuilder struct {
	cfg *JSONConfig
}

// Build creates a map and populates it
func (mb *MapBuilder) Build(ctx context.Context, levelLabel, loggerName, message string) map[string]interface{} {

	outer := make(map[string]interface{})

	for _, f := range mb.cfg.Fields {

		outer[f.Name] = f.generator(ctx, levelLabel, loggerName, message, f)

	}

	return outer

}

// ValueGenerator functions are able to generate a value for a field in a JSON formatted log entry
type ValueGenerator func(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{}

func messageGenerator(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return message
}

func componentGenerator(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return loggerName
}

func levelGenerator(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return levelLabel
}

func utcTimestampGenerator(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return time.Now().UTC().Format(field.Arg)
}

func localTimestampGenerator(ctx context.Context, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return time.Now().UTC().Format(field.Arg)
}
