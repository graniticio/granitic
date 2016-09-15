package ctl

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/logging"
	"sort"
)

type CommandManager struct {
	FrameworkLogger logging.Logger
	commands        map[string]Command
}

func (cm *CommandManager) Find(name string) Command {
	return cm.commands[name]
}

func (cm *CommandManager) Register(command Command) error {

	name := command.Name()

	if cm.commands == nil {
		cm.commands = make(map[string]Command)
	}

	if cm.Find(name) != nil {
		m := fmt.Sprintf("A command named %s is already registered. Command names must be unique.\n", name)
		return errors.New(m)
	}

	cm.commands[name] = command
	cm.FrameworkLogger.LogDebugf("Registered command %s", name)

	return nil
}

func (cm *CommandManager) All() []Command {

	if cm.commands == nil {
		return []Command{}
	}

	s := make([]Command, 0)

	for _, v := range cm.commands {
		s = append(s, v)
	}

	sort.Sort(ByName{s})

	return s
}
