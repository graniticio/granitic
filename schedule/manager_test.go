package schedule

import (
	"github.com/graniticio/granitic/v3/logging"
	"testing"
	"time"
)

func TestInvocationManager(t *testing.T) {

	tsk := new(Task)

	tsk.Name = "mock-task"
	tsk.ID = "id"
	tsk.logic = new(nullLogic)

	fn := tsk.FullName()

	if fn != "mock-task (id)" {
		t.Errorf("Unexpected full task name")
	}

	im := newInvocationManager(tsk)
	im.Log = new(logging.ConsoleErrorLogger)

	iv := &interval{
		Mode:      OffsetFromStart,
		Frequency: time.Hour * 1000,
	}

	im.Interval = iv

	im.determineWait()

}

func TestErrorCreation(t *testing.T) {

	e := NewAllowRetryErrorf("Error = %s", "A")

	a := e.(*AllowRetryError)

	if a.message != "Error = A" {
		t.Error()
	}
}

type nullLogic struct{}

func (nl *nullLogic) ExecuteTask(c chan TaskStatusUpdate) error {
	return nil
}
