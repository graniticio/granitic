package logger

import (
	"github.com/graniticio/granitic/v2/logging"
	"testing"
)

func TestControlLogLevelCommand(t *testing.T) {

	fm := logging.CreateComponentLoggerManager(logging.Fatal, map[string]interface{}{"grncComp": "FATAL"}, []logging.LogWriter{}, nil)

	fm.LoggerByName("grncComp")

	am := new(logging.ComponentLoggerManager)
	am.SetGlobalThreshold(logging.Error)

	am.SetInitialLogLevels(map[string]interface{}{"myComp": "ERROR"})
	am.LoggerByName("myComp")

	lld := new(logLevelCommand)

	lld.FrameworkManager = fm
	lld.ApplicationManager = fm

	lld.ExecuteCommand([]string{}, map[string]string{})
	lld.ExecuteCommand([]string{}, map[string]string{"fw": "true"})

	lld.ExecuteCommand([]string{"myComp", "XXX"}, map[string]string{})
	lld.ExecuteCommand([]string{"myComp", "TRACE"}, map[string]string{})

}
