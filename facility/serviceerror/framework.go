package serviceerror

import (
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

type FrameworkErrorEvent string

const (
	UnableToParseRequest = "UnableToParseRequest"
)

type FrameworkErrorGenerator struct {
	Messages        map[string][]string
	FrameworkLogger logging.Logger
}

func (feg *FrameworkErrorGenerator) Error(e FrameworkErrorEvent, c ws.ServiceErrorCategory, a ...string) *ws.CategorisedError {

	return nil

}

func (feg *FrameworkErrorGenerator) StartComponent() error {

	return nil

}
