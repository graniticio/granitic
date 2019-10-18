package logger

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"testing"
)

func TestNullLoggingComponents(t *testing.T) {

	lm := new(logging.ComponentLoggerManager)
	lm.Disable()

	ca := new(config.Accessor)
	s := new(instance.System)

	cn := ioc.NewComponentContainer(lm, ca, s)

	fb := new(NullLoggingFacilityBuilder)

	fb.BuildAndRegister(lm, ca, cn)

	ald := cn.ProtoComponents()[applicationLoggingDecoratorName]

	if ald == nil {
		t.FailNow()
	}

}
