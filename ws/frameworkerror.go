package ws

import (
	"github.com/graniticio/granitic/logging"
	"fmt"
)

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

func (feg *FrameworkErrorGenerator) Message(e FrameworkErrorEvent,  a ...interface{}) string {

	l := feg.FrameworkLogger
	mc := feg.Messages[e]

	if mc == nil || len(mc) < 2{
		l.LogWarnf("No framework error message defined for '%s'. Returning a default message.")
		return "No error message defined for this error"
	}

	t := mc[1]

	return fmt.Sprintf(t, a...)

}