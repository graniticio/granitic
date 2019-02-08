// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"errors"
	"fmt"
	"strings"
)

// Numeric score of the significance of a message, where zero is non-significant and higher values are more significant.
type LogLevel uint

const (
	All   = 0
	Trace = 10
	Debug = 20
	Info  = 40
	Warn  = 50
	Error = 60
	Fatal = 70
)

const (
	AllLabel   = "ALL"
	TraceLabel = "TRACE"
	DebugLabel = "DEBUG"
	InfoLabel  = "INFO"
	WarnLabel  = "WARN"
	ErrorLabel = "ERROR"
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
