package runtimectl

import (
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
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
