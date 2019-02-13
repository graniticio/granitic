package ctl

import (
	"github.com/graniticio/granitic/v2/ioc"
	"testing"
)

func TestDecorator(t *testing.T) {

	m := createManager()

	cd := new(CommandDecorator)
	cd.CommandManager = m
	cd.FrameworkLogger = m.FrameworkLogger

	c := ioc.NewComponent("mock-comm", new(mockCommand))

	if !cd.OfInterest(c) {
		t.Fatal()
	}

	cd.DecorateComponent(c, nil)

	if len(m.commands) == 0 {
		t.Error()
	}

}
