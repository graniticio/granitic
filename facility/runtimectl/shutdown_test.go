package runtimectl

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"testing"
)

func TestShutdownCommand(t *testing.T) {

	sc := new(shutdownCommand)
	sc.FrameworkLogger = new(logging.ConsoleErrorLogger)
	sc.disableExit = true

	fm := logging.CreateComponentLoggerManager(logging.Fatal, map[string]interface{}{"grncComp": "FATAL"}, []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	cc := ioc.NewComponentContainer(fm, new(config.Accessor), new(instance.System))
	sc.container = cc

	sc.ExecuteCommand([]string{}, map[string]string{})

}
