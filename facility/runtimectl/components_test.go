package runtimectl

import (
	"github.com/graniticio/granitic/v2/ioc"
	"testing"
)

func TestContainerInjection(t *testing.T) {

	cc := new(componentsCommand)
	cc.Container(new(ioc.ComponentContainer))

}
