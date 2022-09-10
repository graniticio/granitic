package logger

import (
	"github.com/graniticio/granitic/v3/logging"
	"testing"
)

func TestGlobalLevelCommand(t *testing.T) {

	fm := new(logging.ComponentLoggerManager)
	fm.SetGlobalThreshold(logging.Fatal)

	am := new(logging.ComponentLoggerManager)
	am.SetGlobalThreshold(logging.Error)

	glc := new(globalLogLevelCommand)

	glc.FrameworkLogger = new(logging.ConsoleErrorLogger)

	glc.ApplicationManager = am
	glc.FrameworkManager = fm

	glc.ExecuteCommand([]string{"TRACE"}, map[string]string{"fw": "true"})

	if fm.GlobalLevel() != logging.Trace {
		t.Errorf("Expected framework level to be at TRACE")
	}

	glc.ExecuteCommand([]string{"TRACE"}, map[string]string{})

	if am.GlobalLevel() != logging.Trace {
		t.Errorf("Expected application level to be at TRACE")
	}

	glc.ExecuteCommand([]string{}, map[string]string{"fw": "true"})
	glc.ExecuteCommand([]string{}, map[string]string{})

}
