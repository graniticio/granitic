package runtimectl

import (
	"github.com/graniticio/granitic/v3/ctl"
	"testing"
)

func TestHelpCommand(t *testing.T) {

	hc := new(helpCommand)
	hc.commandManager = new(ctl.CommandManager)

	hc.ExecuteCommand([]string{}, map[string]string{})

}
