package schedule

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"time"
)

const firstRunFormat = "2006-01-02 15:04:05"

func newInvocationManager(t *Task) *invocationManager {

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

	task := im.Task
	im.State = ioc.StartingState

	if !task.Disabled {
		im.setFirstInvocation()
	}

	im.State = ioc.RunningState

	for im.State != ioc.StoppedState && im.State != ioc.StoppingState {

		// The first invocation on the queue is the next to be run
		next := im.scheduled.PeekHead()

		if next != nil {
			runAt := next.runAt

			now := time.Now()

			if runAt == now || runAt.Before(now) {
				// The time at which this invocation was scheduled to run has arrived (or passed)
				im.scheduled.Dequeue()

				if next.reason == Scheduled && !task.Disabled {
					im.addNextInvocation(next)
				}

				// Only run this invocation if this task manager is running
				if im.State == ioc.RunningState {

					if im.running.Size() != 0 && !task.NoWarnOnOverlap {
						im.Log.LogWarnf("%d instance(s) of task %s are already running.", im.running.Size(), im.Task.FullName())
					}

					if im.running.Size() > uint64(task.MaxOverlapping) {
						im.Log.LogErrorf("Will not start new instance of task %s as %d instance(s) are already running - maximum allowed is %d", im.Task.FullName(), im.running.Size(), im.Task.MaxOverlapping+1)
					} else {

						im.running.EnqueueAtTail(next)

						go im.runTask(next)
					}
				}

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

		m := fmt.Sprintf("Problem executing task %s (invocation %d, attempt %d started at %v): %s", im.Task.FullName(), i.counter, i.attempt, i.startedAt, err.Error())

		if _, ok := err.(*AllowRetryError); ok {

			if okay, when := im.attemptRetry(i); okay {
				im.Log.LogWarnf(m)
				im.Log.LogWarnf("Will retry at %v", when)
			} else {
				im.Log.LogErrorf(m)
			}

		} else {
			im.Log.LogErrorf(m)
		}

	}

}

// See if the invocation of a task can be tried again
func (im *invocationManager) attemptRetry(i *invocation) (bool, time.Time) {

	if !i.retryAllowed() {
		return false, time.Now()
	}

	retryTime := time.Now().Add(im.Task.retryWait)

	nextScheduled := im.scheduled.PeekHead().runAt

	if nextScheduled.Before(retryTime) {
		//No point retrying as next scheduled run will happen before that
		im.Log.LogWarnf("Retry attempt attempt abandoned as next scheduled invocation will arrive first")
		return false, time.Now()
	}

	i.attempt++
	i.runAt = retryTime
	i.reason = Retry

	im.scheduled.EnqueueAtHead(i)

	return true, retryTime

}

func (im *invocationManager) listenForStatusUpdates(i *invocation, ch chan TaskStatusUpdate) {

	task := im.Task

	ts := TaskInvocationSummary{
		InvocationCount: i.counter,
		StartedAt:       i.startedAt,
		TaskID:          task.ID,
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
	}

	untilNext := next.runAt.Sub(time.Now())

	if untilNext < time.Second {
		return untilNext
	}

	return time.Second
}

func (im *invocationManager) setFirstInvocation() {

	interval := im.Interval

	i := newInvocation(1, im.Task.MaxRetries, Scheduled)

	if interval.Mode == OffsetFromStart {
		i.runAt = time.Now().Add(interval.OffsetFromStart)
	} else {
		i.runAt = interval.ActualStart
	}

	im.Log.LogInfof("Task '%s' will first run at %s and intervals of %v thereafter", im.Task.FullName(), i.runAt.Format(firstRunFormat), interval.Frequency)

	t := im.Task

	if t.MaxRetries > 0 {
		im.Log.LogDebugf("Task '%s' can be retried a maximum of %d time(s) with an interval of %v between retries", t.FullName(), t.MaxRetries, t.retryWait)
	}

	im.scheduled.EnqueueAtTail(i)

}

func (im *invocationManager) addNextInvocation(previous *invocation) time.Time {

	interval := im.Interval

	i := newInvocation(previous.counter+1, im.Task.MaxRetries, Scheduled)
	i.runAt = previous.runAt.Add(interval.Frequency)

	im.scheduled.EnqueueAtTail(i)

	return i.runAt

}

func (im *invocationManager) PrepareToStop() {
	im.State = ioc.StoppingState
}

func (im *invocationManager) ReadyToStop() (bool, error) {

	if im.running.Size() == 0 {
		return true, nil
	}

	return false, fmt.Errorf("%d instance(s) of task %s are running", im.running.size, im.Task.FullName())

}

func (im *invocationManager) Stop() error {
	if im.running.Size() > 0 {
		m := fmt.Sprintf("%d instance(s) of task %s are still running", im.running.size, im.Task.FullName())
		return errors.New(m)
	}

	return nil
}
