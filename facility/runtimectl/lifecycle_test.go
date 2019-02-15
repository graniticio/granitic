package runtimectl

import (
	"github.com/graniticio/granitic/v2/ioc"
	"testing"
)

func TestLifecycleCommand(t *testing.T) {

	lc := new(lifecycleCommand)
	lc.filterFunc = fc
	lc.ExecuteCommand([]string{}, map[string]string{})

}

func fc(*ioc.ComponentContainer, bool, ...string) []*ioc.Component {
	return []*ioc.Component{}
}
