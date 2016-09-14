package ctl

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type CommandDecorator struct {
	FrameworkLogger logging.Logger
	CommandManager  *CommandManager
}

func (cd *CommandDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(Command)

	return found

}

func (cd *CommandDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	c := component.Instance.(Command)

	if err := cd.CommandManager.Register(c); err != nil {

		cd.FrameworkLogger.LogErrorf(err.Error())

	}
}
