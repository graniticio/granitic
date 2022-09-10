package runtimectl

import (
	"github.com/graniticio/granitic/v3/ioc"
	"testing"
)

func TestContainerInjection(t *testing.T) {

	cc := new(componentsCommand)
	cc.Container(new(ioc.ComponentContainer))

}
