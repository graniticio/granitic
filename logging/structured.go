// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/types"
	"strings"
)

// A JSONLogFormatter is a component able to take a message to be written to a log file and format it as JSON document
type JSONLogFormatter struct {
	// A component able to extract information from a context.Context into a loggable format
	ContextFilter ContextFilter
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (jlf *JSONLogFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {
	return fmt.Sprintf("{%s %s}", loggerName, message)
}

//SetContextFilter provides the formatter with access selected data from a context
func (jlf *JSONLogFormatter) SetContextFilter(cf ContextFilter) {
	jlf.ContextFilter = cf
}

// JSONConfig defines the fields to be included in a  JSON-formatted application log entry
type JSONConfig struct {
	Prefix string
	Fields []JSONField
	Suffix string
}

// A JSONField defines the rules for outputting a single field in a JSON-formatted application log entry
type JSONField struct {
	Name    string
	Content string
	Arg     string
}

const (
	message   = "MESSAGE"
	comp      = "COMPONENT_NAME"
	timestamp = "TIMESTAMP"
	level     = "LEVEL"
	ctxVal    = "CONTEXT_VALUE"
)

// ValidateJSONFields checks that the configuration of a JSON application log entry is correct
func ValidateJSONFields(fields []JSONField) error {

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
				return fmt.Errorf("you must specify an Arg when using JSON fields with the content type %s (a standard Go date/time format string)", timestamp)
			}

		}

	}

	return nil

}
