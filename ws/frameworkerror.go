package ws

import (
	"github.com/graniticio/granitic/logging"
	"fmt"
)

type WsFrameworkPhase int

const (
	Unmarshall = iota
	QueryBind
	PathBind
)

type WsFrameworkError struct {
	Phase       WsFrameworkPhase
	ClientField string
	TargetField string
	Message     string
	Value string
	Position int
	Code string
}

func NewUnmarshallWsFrameworkError(message, code string) *WsFrameworkError {
	f := new(WsFrameworkError)
	f.Phase = Unmarshall
	f.Message = message
	f.Code = code

	return f
}

func NewQueryBindFrameworkError(message, code, param, target string) *WsFrameworkError {
	f := new(WsFrameworkError)
	f.Phase = QueryBind
	f.Message = message
	f.ClientField = param
	f.TargetField = target
	f.Code = code

	return f
}

func NewPathBindFrameworkError(message, code, target string) *WsFrameworkError {
	f := new(WsFrameworkError)
	f.Phase = PathBind
	f.Message = message
	f.TargetField = target
	f.Code = code

	return f
}


type FrameworkErrorEvent string

const (
	UnableToParseRequest = "UnableToParseRequest"
	QueryTargetNotArray = "QueryTargetNotArray"
	QueryWrongType = "QueryWrongType"
	PathWrongType = "PathWrongType"
)

type FrameworkErrorGenerator struct {
	Messages        map[FrameworkErrorEvent][]string
	FrameworkLogger logging.Logger
}

func (feg *FrameworkErrorGenerator) Error(e FrameworkErrorEvent, c ServiceErrorCategory, a ...interface{}) *CategorisedError {

	l := feg.FrameworkLogger
	mc := feg.Messages[e]

	if mc == nil || len(mc) < 2{
		l.LogWarnf("No framework error message defined for '%s'. Returning a default message.")
		return NewCategorisedError(c, "UNKNOWN", "No error message defined for this error")
	}

	cd := mc[0]
	t := mc[1]

	fm := fmt.Sprintf(t, a...)

	return NewCategorisedError(c, cd, fm)

}

func (feg *FrameworkErrorGenerator) MessageCode(e FrameworkErrorEvent,  a ...interface{}) (message string, code string) {

	l := feg.FrameworkLogger
	mc := feg.Messages[e]

	if mc == nil || len(mc) < 2{
		l.LogWarnf("No framework error message defined for '%s'. Returning a default message.")
		return "No error message defined for this error", "UNKNOWN"
	}

	t := mc[1]

	return fmt.Sprintf(t, a...), mc[0]

}
