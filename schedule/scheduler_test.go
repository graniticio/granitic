package schedule

import (
	"github.com/graniticio/granitic/v2/ioc"
	"testing"
)

func TestLifecycleImplementations(t *testing.T) {

	var _ ioc.Startable = (*TaskScheduler)(nil)
	var _ ioc.Stoppable = (*TaskScheduler)(nil)

}
