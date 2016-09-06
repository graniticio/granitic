package logging

import (
	"errors"
	"fmt"
	"strings"
)

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
	TraceLabel = "TRACE"
	DebugLabel = "DEBUG"
	InfoLabel  = "INFO"
	WarnLabel  = "WARN"
	ErrorLabel = "ERROR"
	FatalLabel = "FATAL"
)

func LogLevelFromLabel(label string) (LogLevel, error) {

	u := strings.ToUpper(label)

	switch u {
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

	m := fmt.Sprintf("%s is not valid log level. Valid levels are %s, %s, %s, %s, %s, %s.",
		u, TraceLabel, DebugLabel, InfoLabel, WarnLabel, ErrorLabel, FatalLabel)

	return All, errors.New(m)
}
