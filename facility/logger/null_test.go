package logger

import (
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
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
