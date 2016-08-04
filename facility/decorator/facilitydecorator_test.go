package decorator

import (
	"testing"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/test"
)

type AcceptLog struct {
	Log logging.Logger
}

type RejectLogStruct struct {
	Log map[string]string
}

type RejectLogPrim struct {
	Log int
}


func TestMatcher(t *testing.T) {

	d := new(ApplicationLogDecorator)
	d.FrameworkLogger = logging.CreateAnonymousLogger("test", logging.All)

	c := new(ioc.Component)
	c.Name = "Match"
	c.Instance = new(AcceptLog)

	test.ExpectBool(t, d.OfInterest(c), true)

	c.Name = "NoMatch"
	c.Instance = new (RejectLogStruct)

	test.ExpectBool(t, d.OfInterest(c), false)

	c.Name = "NoMatchPrim"
	c.Instance = new (RejectLogPrim)

	test.ExpectBool(t, d.OfInterest(c), false)

}
