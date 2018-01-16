package schedule

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"time"
)

func NewInvocationManager(t *Task) *invocationManager {

	im := new(invocationManager)
	im.Task = t
	im.scheduled = new(invocationQueue)
	im.running = new(invocationQueue)
	im.State = ioc.StoppedState
	return im

}

type invocationManager struct {
	Task      *Task
	Interval  *interval
	scheduled *invocationQueue
	running   *invocationQueue
	State     ioc.ComponentState
	Log       logging.Logger
}

func (im *invocationManager) Start() {

	im.State = ioc.StartingState

	im.setFirstInvocation()

	im.State = ioc.RunningState

	for im.State != ioc.StoppedState && im.State != ioc.StoppingState {
		//Pop tasks from the schedule queue as long as we're not stopping

		first := im.scheduled.Peek()

		if first != nil {
			runAt := first.runAt

			now := time.Now()

			if runAt == now || runAt.Before(now) {
				im.scheduled.Dequeue()

				if im.State == ioc.RunningState {
					im.running.Enqueue(first)

					go im.runTask(first)
				}

				im.addNextInvocation(first)
			}

		}

		waitTime := im.determineWait()

		time.Sleep(waitTime)
	}
}

func (im *invocationManager) runTask(i *invocation) {

	i.startedAt = time.Now()

	if im.Log.IsLevelEnabled(logging.Trace) {
		im.Log.LogTracef("Accuracy: %v", i.startedAt.Sub(i.runAt))
	}

	updates := make(chan TaskStatusUpdate)

	im.Log.LogDebugf("Executing %s", im.Task.FullName())

	defer func() {
		if r := recover(); r != nil {
			im.Log.LogErrorfWithTrace("Panic recovered while executing task %s (invocation %d started at %v)\n %v", im.Task.FullName(), i.counter, i.startedAt, r)
		}
		close(updates)
	}()

	err := im.Task.logic.ExecuteTask(updates)

	if err != nil {
		im.Log.LogErrorf("Problem executing task %s (invocation %d started at %v): %s", im.Task.FullName(), i.counter, i.startedAt, err.Error())
	}

	close(updates)
}

func (im *invocationManager) determineWait() time.Duration {

	next := im.scheduled.Peek()

	if next == nil {
		return time.Second
	} else {

		untilNext := next.runAt.Sub(time.Now())

		if untilNext < time.Second {
			return untilNext
		} else {
			return time.Second
		}

	}

}

func (im *invocationManager) setFirstInvocation() {

	interval := im.Interval

	i := new(invocation)
	i.counter = 1

	if interval.Mode == OFFSET_FROM_START {
		i.runAt = time.Now().Add(interval.OffsetFromStart)
	} else {
		i.runAt = interval.ActualStart
	}

	im.scheduled.Enqueue(i)

}

func (im *invocationManager) addNextInvocation(previous *invocation) time.Time {

	interval := im.Interval

	i := new(invocation)
	i.counter = previous.counter + 1
	i.runAt = previous.runAt.Add(interval.Frequency)

	im.scheduled.Enqueue(i)

	return i.runAt

}

func (im *invocationManager) PrepareToStop() {
	im.State = ioc.StoppingState
}
