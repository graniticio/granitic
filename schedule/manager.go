package schedule

import (
	"fmt"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"errors"
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
		first := im.scheduled.PeekHead()

		if first != nil {
			runAt := first.runAt

			now := time.Now()

			if runAt == now || runAt.Before(now) {
				im.scheduled.Dequeue()

				if im.State == ioc.RunningState {

					if im.running.Size() != 0 && !im.Task.NoWarnOnOverlap {
						im.Log.LogWarnf("%d instance(s) of task %s are already running.", im.running.Size(), im.Task.FullName())
					}

					if im.running.Size() > uint64(im.Task.MaxOverlapping) {
						im.Log.LogErrorf("Will not start new instance of task %s as %d instance(s) are already running - maximum allowed is %d", im.Task.FullName(), im.running.Size(), im.Task.MaxOverlapping+1)
					} else {

						im.running.Enqueue(first)

						go im.runTask(first)
					}
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

	updates := make(chan TaskStatusUpdate, 20)

	im.Log.LogDebugf("Executing %s", im.Task.FullName())

	if im.Task.LogStatusMessages || im.Task.receiver != nil {
		go im.listenForStatusUpdates(i, updates)
	}

	defer func() {
		if r := recover(); r != nil {
			im.Log.LogErrorfWithTrace("Panic recovered while executing task %s (invocation %d started at %v)\n %v", im.Task.FullName(), i.counter, i.startedAt, r)
		}

		close(updates)
		im.running.Remove(i.counter)

	}()

	err := im.Task.logic.ExecuteTask(updates)

	if err != nil {
		im.Log.LogErrorf("Problem executing task %s (invocation %d started at %v): %s", im.Task.FullName(), i.counter, i.startedAt, err.Error())
	}

}

func (im *invocationManager) listenForStatusUpdates(i *invocation, ch chan TaskStatusUpdate) {

	task := im.Task

	ts := TaskInvocationSummary{
		InvocationCount: i.counter,
		StartedAt:       i.startedAt,
		TaskId:          task.Id,
		TaskName:        task.Name,
	}

	for {
		su, ok := <-ch

		if !ok {
			break
		}

		if task.LogStatusMessages && len(su.Message) > 0 {
			im.Log.LogInfof("Task: %s Invocation: %d: %s", task.FullName(), i.counter, su.Message)
		}

		if task.receiver != nil {
			task.receiver.Receive(ts, su)
		}

	}

}

func (im *invocationManager) determineWait() time.Duration {

	next := im.scheduled.PeekHead()

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

	if im.Log.IsLevelEnabled(logging.Debug) {
		im.Log.LogDebugf("Task %s will first run at %v and intervals of %v thereafter", im.Task.FullName(), i.runAt, interval.Frequency)
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

func (im *invocationManager) ReadyToStop() (bool, error) {

	if im.running.Size() == 0 {
		return true, nil
	} else {

		m := fmt.Sprintf("%d instance(s) of task %s are running", im.running.size, im.Task.FullName())
		return false, errors.New(m)
	}

}

func (im *invocationManager) Stop() error {
	if im.running.Size() > 0 {
		m := fmt.Sprintf("%d instance(s) of task %s are still running", im.running.size, im.Task.FullName())
		return errors.New(m)
	}

	return nil
}
