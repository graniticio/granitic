package schedule

import (
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"testing"
)

func TestLifecycleImplementations(t *testing.T) {

	var _ ioc.Startable = (*TaskScheduler)(nil)
	var _ ioc.Stoppable = (*TaskScheduler)(nil)

}

func TestSchedulerSetup(t *testing.T) {

	ts := new(TaskScheduler)

	ts.FrameworkLogger = new(logging.ConsoleErrorLogger)

	tsk := new(Task)

	tsk.Name = "mock-task"
	tsk.ID = "id"
	tsk.logic = new(nullLogic)

	con := &mockSingleTaskContainer{
		t:  tsk,
		cn: "taskCompName",
	}

	ts.componentContainer = con

	ts.StartComponent()
	ts.AllowAccess()
	ts.PrepareToStop()
	ts.Stop()

}

type mockSingleTaskContainer struct {
	t  *Task
	cn string
}

func (c *mockSingleTaskContainer) ComponentByName(string) *ioc.Component {
	return ioc.NewComponent(c.cn, c.t)
}

func (c *mockSingleTaskContainer) AllComponents() []*ioc.Component {

	return []*ioc.Component{ioc.NewComponent(c.cn, c.t)}

}
