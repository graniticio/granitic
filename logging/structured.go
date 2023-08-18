// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/types"
)

// A JSONLogFormatter is a component able to take a message to be written to a log file and format it as JSON document
type JSONLogFormatter struct {
	Config     *JSONConfig
	MapBuilder *MapBuilder
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (jlf *JSONLogFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {

	m := jlf.MapBuilder.Build(ctx, levelLabel, loggerName, message)
	cfg := jlf.Config

	entry, _ := json.Marshal(m)

	return cfg.Prefix + string(entry) + cfg.Suffix
}

// AllowAccess checks that a context filter has been injected (if the field configuration needs on)
func (jlf *JSONLogFormatter) AllowAccess() error {

	mb := jlf.MapBuilder

	if mb.RequiresContextFilter && mb.contextFilter == nil {
		return fmt.Errorf("your JSON application logging configuration includes fields that display information from the context, but no component is available that implements logging.ContextFilter")
	}

	return nil
}

// SetInstanceID accepts the current instance ID
func (jlf *JSONLogFormatter) SetInstanceID(i *instance.Identifier) {
	if jlf.MapBuilder != nil {
		jlf.MapBuilder.instanceID = i
	}
}

// SetContextFilter provides the formatter with access selected data from a context
func (jlf *JSONLogFormatter) SetContextFilter(cf ContextFilter) {
	jlf.MapBuilder.contextFilter = cf
}

// JSONConfig defines the fields to be included in a  JSON-formatted application log entry
type JSONConfig struct {
	Prefix       string
	Fields       [][]string
	ParsedFields []*JSONField
	Suffix       string
	UTC          bool
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
	firstLine = "FIRST_LINE"
	skipFirst = "SKIP_FIRST"
	comp      = "COMPONENT_NAME"
	timestamp = "TIMESTAMP"
	level     = "LEVEL"
	ctxVal    = "CONTEXT_VALUE"
	text      = "TEXT"
	inst      = "INSTANCE_ID"
	levelMap  = "LEVEL_MAP"
)

// ConvertFields converts from the config representation of a field list to the internal version
func ConvertFields(unparsed [][]string) []*JSONField {

	l := len(unparsed)

	if l == 0 {
		return make([]*JSONField, 0)
	}

	allParsed := make([]*JSONField, l)

	for i, raw := range unparsed {

		parsed := new(JSONField)
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

// ValidateJSONFields checks that the configuration of a JSON application log entry is correct
func ValidateJSONFields(fields []*JSONField) error {

	seenLevelMap := false

	allowed := types.NewOrderedStringSet([]string{message, comp, timestamp, level, ctxVal, firstLine, skipFirst, text, inst, levelMap})

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

		if f.Content == levelMap {

			if seenLevelMap {
				return fmt.Errorf("you can only have one JSON application logging field of type %s", levelMap)
			}

			if _, err := parseLevelMap(f.Arg); err != nil {
				return nil
			}
		}

	}

	return nil

}

func parseLevelMap(format string) (map[string]string, error) {

	split := strings.Split(format, ",")
	m := make(map[string]string, len(split))

	for _, mapping := range split {

		kv := strings.Split(mapping, ":")

		if len(kv) != 2 {
			return nil, fmt.Errorf("%s does not appear to a valid mapping (should be of the form GRANTIC_LEVEL:MAPPED_LEVEL", mapping)
		}

		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])

		if _, present := m[k]; present {
			return nil, fmt.Errorf("Multiple mappings found for log level %s", k)
		}

		m[k] = v

	}

	return m, nil
}

// CreateMapBuilder builds a component able to generate a log entry based on the rules in the supplied fields.
func CreateMapBuilder(cfg *JSONConfig) (*MapBuilder, error) {

	mb := new(MapBuilder)

	mb.cfg = cfg

	for _, f := range cfg.ParsedFields {

		switch f.Content {
		case message:
			f.generator = mb.messageGenerator
		case comp:
			f.generator = mb.componentGenerator
		case level:
			f.generator = mb.levelGenerator
		case timestamp:
			if cfg.UTC {
				f.generator = mb.utcTimestampGenerator
			} else {
				f.generator = mb.localTimestampGenerator
			}
		case ctxVal:
			mb.RequiresContextFilter = true
			f.generator = mb.ctxValGenerator
		case text:
			f.generator = mb.textGenerator
		case firstLine:
			f.generator = mb.firstLineGenerator
		case skipFirst:
			f.generator = mb.skipFirstGenerator
		case inst:
			f.generator = mb.instanceIDGenerator
		case levelMap:
			mb.levelMap, _ = parseLevelMap(f.Arg)
			f.generator = mb.levelMapGenerator

		}
	}

	return mb, nil
}

// MapBuilder creates a map[string]interface{} representing a log entry, ready for JSON encoding
type MapBuilder struct {
	cfg                   *JSONConfig
	contextFilter         ContextFilter
	RequiresContextFilter bool
	instanceID            *instance.Identifier
	levelMap              map[string]string
}

// Build creates a map and populates it
func (mb *MapBuilder) Build(ctx context.Context, levelLabel, loggerName, message string) map[string]interface{} {

	var fcd FilteredContextData

	outer := make(map[string]interface{})

	if mb.RequiresContextFilter && mb.contextFilter != nil && ctx != nil {
		fcd = mb.contextFilter.Extract(ctx)
	}

	for _, f := range mb.cfg.ParsedFields {

		outer[f.Name] = f.generator(fcd, levelLabel, loggerName, message, f)

	}

	return outer

}

// ValueGenerator functions are able to generate a value for a field in a JSON formatted log entry
type ValueGenerator func(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{}

func (mb *MapBuilder) messageGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return message
}

func (mb *MapBuilder) firstLineGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return strings.Split(message, "\n")[0]
}

func (mb *MapBuilder) skipFirstGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {

	l := strings.SplitAfterN(message, "\n", 2)

	if len(l) != 2 {
		return ""
	}

	return l[1]
}

func (mb *MapBuilder) componentGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return loggerName
}

func (mb *MapBuilder) textGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return field.Arg
}

func (mb *MapBuilder) levelGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return levelLabel
}

func (mb *MapBuilder) utcTimestampGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return time.Now().UTC().Format(field.Arg)
}

func (mb *MapBuilder) localTimestampGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	return time.Now().UTC().Format(field.Arg)
}

func (mb *MapBuilder) ctxValGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {

	if fcd != nil {
		return fcd[field.Arg]
	}

	return ""
}

func (mb *MapBuilder) instanceIDGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {
	if mb.instanceID != nil {
		return mb.instanceID.ID
	}

	return ""
}

func (mb *MapBuilder) levelMapGenerator(fcd FilteredContextData, levelLabel, loggerName, message string, field *JSONField) interface{} {

	lm := mb.levelMap

	if lm == nil {
		return levelLabel
	}

	if mapped, okay := lm[levelLabel]; okay {
		return mapped
	}

	return levelLabel

}
