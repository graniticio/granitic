package ctl

import (
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"github.com/graniticio/granitic/v2/ws"
	"testing"
)

func TestRegistration(t *testing.T) {

	m := createManager()

	err := m.Register(new(mockCommand))

	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

}

func createManager() *CommandManager {

	m := new(CommandManager)

	m.commands = make(map[string]Command)
	m.Disabled = []string{}
	m.DisabledLookup = types.NewEmptyOrderedStringSet()
	m.FrameworkLogger = new(logging.ConsoleErrorLogger)

	return m

}

type mockCommand struct {
}

func (mc *mockCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*CommandOutput, []*ws.CategorisedError) {

	co := new(CommandOutput)

	co.OutputHeader = "head"
	co.OutputBody = [][]string{{"line1A", "line1B"}, {"line2A", "line2B"}}
	co.RenderHint = Columns

	return co, []*ws.CategorisedError{}
}

func (mc *mockCommand) Name() string {
	return "mock"
}

func (mc *mockCommand) Summmary() string {
	return "summary"
}

func (mc *mockCommand) Usage() string {
	return "usage"
}

func (mc *mockCommand) Help() []string {
	return []string{"help1", "help2"}
}
