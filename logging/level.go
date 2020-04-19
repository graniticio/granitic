// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"errors"
	"fmt"
	"strings"
)

// LogLevel is the numeric score of the significance of a message, where zero is non-significant and higher values are more significant.
type LogLevel uint

const (
	// All allows all messages to be logged
	All = 0
	// Trace allows messages with a significance of Trace or higher to be logged
	Trace = 10
	// Debug allows messages with a significance of Debug or higher to be logged
	Debug = 20
	// Info allows messages with a significance of Info or higher to be logged
	Info = 40
	// Warn allows messages with a significance of Warn or higher to be logged
	Warn = 50
	// Error allows messages with a significance of Error or higher to be logged
	Error = 60
	// Fatal allows messages with a significance of Fatal or higher to be logged
	Fatal = 70
)

const (
	//AllLabel maps the string ALL to the numeric log level All
	AllLabel = "ALL"
	//TraceLabel maps the string TRACE to the numeric log level Trace
	TraceLabel = "TRACE"
	//DebugLabel maps the string DEBUG to the numeric log level Debug
	DebugLabel = "DEBUG"
	//InfoLabel maps the string INFO to the numeric log level Debug
	InfoLabel = "INFO"
	//WarnLabel maps the string WARN to the numeric log level Warn
	WarnLabel = "WARN"
	//ErrorLabel maps the string ERROR to the numeric log level Error
	ErrorLabel = "ERROR"
	//FatalLabel maps the string FATAL to the numeric log level Fatal
	FatalLabel = "FATAL"
)

// LogLevelFromLabel takes a string name for a log level (TRACE, DEBUG etc) and finds a numeric threshold associated with
// that type of message.
func LogLevelFromLabel(label string) (LogLevel, error) {

	u := strings.ToUpper(label)

	switch u {
	case AllLabel:
		return All, nil
	case WarnLabel:
		return Warn, nil
	case TraceLabel:
		return Trace, nil
	case DebugLabel:
		return Debug, nil
	case InfoLabel:
		return Info, nil
	case ErrorLabel:
		return Error, nil
	case FatalLabel:
		return Fatal, nil
	}

	m := invalidLogLevelMessage(label)

	return All, errors.New(m)
}

// LabelFromLevel takes a member of the LogLevel enumeration (All, Fatal) and converts it to a string code ('ALL', 'FATAL').
// If the supplied LogLevel cannot be mapped to a defined level, the word 'CUSTOM' is returned.
func LabelFromLevel(ll LogLevel) string {

	switch ll {
	case All:
		return AllLabel
	case Trace:
		return TraceLabel
	case Debug:
		return DebugLabel
	case Info:
		return InfoLabel
	case Warn:
		return WarnLabel
	case Error:
		return ErrorLabel
	case Fatal:
		return FatalLabel
	default:
		return "CUSTOM"
	}

}

func invalidLogLevelMessage(label string) string {

	return fmt.Sprintf("%s is not valid log level. Valid levels are %s, %s, %s, %s, %s, %s, %s.",
		label, AllLabel, TraceLabel, DebugLabel, InfoLabel, WarnLabel, ErrorLabel, FatalLabel)
}
