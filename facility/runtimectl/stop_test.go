package runtimectl

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"testing"
)

func TestStopCommand(t *testing.T) {

	sc := newStopCommand()
	sc.FrameworkLogger = new(logging.ConsoleErrorLogger)

	fm := logging.CreateComponentLoggerManager(logging.Fatal, map[string]interface{}{"grncComp": "FATAL"}, []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter())

	cc := ioc.NewComponentContainer(fm, new(config.Accessor), new(instance.System))
	sc.container = cc

	sc.ExecuteCommand([]string{}, map[string]string{})

}
