// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ctl

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"sort"
)

// A CommandManager is created as a component as part of the RuntimeCtl facility, CommandManager acts a registry for all components that
// implement ctl.Command. See http://granitic.io/1.0/ref/runtime-control for details on how to configure this component.
type CommandManager struct {
	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger
	commands        map[string]Command

	// A slice of command names that should not be registered, effectively disabling their use. Populated via configuration.
	Disabled []string

	// The contents of Disabled converted to a Set at facility instantiation time.
	DisabledLookup types.StringSet
}

// Find returns a command that has a Name() exactly matching the supplied string.
func (cm *CommandManager) Find(name string) Command {
	return cm.commands[name]
}

// Register stores a reference to the supplied Command and will return it via the Find method, unless the name
// of the command is in the Disabled slice. Returns an error if the command's name is already in use.
func (cm *CommandManager) Register(command Command) error {

	name := command.Name()

	if cm.commands == nil {
		cm.commands = make(map[string]Command)
	}

	if cm.Find(name) != nil {
		m := fmt.Sprintf("A command named %s is already registered. Command names must be unique.\n", name)
		return errors.New(m)
	}

	if cm.DisabledLookup.Contains(name) {
		cm.FrameworkLogger.LogDebugf("Ignoring disabled command %s", name)
		return nil
	}

	cm.commands[name] = command
	cm.FrameworkLogger.LogDebugf("Registered command %s", name)

	return nil
}

// All returns a slice of all of the currently registered commands.
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
