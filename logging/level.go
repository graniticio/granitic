package logging

import "strings"

const (
	All   = 0
	Trace = 10
	Debug = 20
	Info  = 40
	Warn  = 50
	Error = 60
	Fatal = 70
)

const TraceLabel = "TRACE"
const DebugLabel = "DEBUG"
const InfoLabel = "INFO"
const WarnLabel = "WARN"
const ErrorLabel = "ERROR"
const FatalLabel = "FATAL"

func LogLevelFromLabel(label string) int {
	switch strings.ToUpper(label) {
	case WarnLabel:
		return Warn
	case TraceLabel:
		return Trace
	case DebugLabel:
		return Debug
	case InfoLabel:
		return Info
	case ErrorLabel:
		return Error
	case FatalLabel:
		return Fatal
	}

	return All
}
