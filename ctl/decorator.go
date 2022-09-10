// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ctl

import (
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
)

// CommandDecorator is used by the Granitic IoC container to discover components implementing ctl.Command and registering them with an instance
// of ctl.CommandManager.
type CommandDecorator struct {
	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger

	// An instance of CommandManager to register discovered Commands with.
	CommandManager *CommandManager
}

// OfInterest checks to see if the supplied component implements ctl.Command.
func (cd *CommandDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(Command)

	return found

}

// DecorateComponent registers the component with the CommandManager.
func (cd *CommandDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	c := component.Instance.(Command)

	if err := cd.CommandManager.Register(c); err != nil {

		cd.FrameworkLogger.LogErrorf(err.Error())

	}
}
