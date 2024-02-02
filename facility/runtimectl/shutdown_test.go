package runtimectl

import (
	config_access "github.com/graniticio/config-access"
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

	cc := ioc.NewComponentContainer(fm, config_access.NewGraniticSelector(make(config_access.ConfigNode)), new(instance.System))
	sc.container = cc

	sc.ExecuteCommand([]string{}, map[string]string{})

}
