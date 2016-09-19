package ioc

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/logging"
	"os"
	"time"
)

type LifecycleSupport int

const (
	None = iota
	CanStart
	CanStop
	CanSuspend
	CanBlockStart
	CanBeAccessed
)

type Startable interface {
	StartComponent() error
}

type Suspendable interface {
	Suspend() error
	Resume() error
}

type Stoppable interface {
	PrepareToStop()
	ReadyToStop() (bool, error)
	Stop() error
}

type AccessibilityBlocker interface {
	BlockAccess() (bool, error)
}

type Accessible interface {
	AllowAccess() error
}

type LifecycleManager struct {
	container       *ComponentContainer
	FrameworkLogger logging.Logger
}

func (lm *LifecycleManager) StartAll() error {

	defer func() {
		if r := recover(); r != nil {
			lm.FrameworkLogger.LogErrorfWithTrace("Panic recovered while starting components components %s", r)
			os.Exit(-1)
		}
	}()

	for _, component := range lm.container.byLifecycleSupport[CanStart] {

		startable := component.Instance.(Startable)

		if err := startable.StartComponent(); err != nil {
			message := fmt.Sprintf("Unable to start %s: %s", component.Name, err)
			return errors.New(message)
		}

	}

	if len(lm.container.byLifecycleSupport[CanBlockStart]) != 0 {
		if err := lm.waitForBlockers(5*time.Second, 12, 0); err != nil {
			return err
		}

	}

	for _, component := range lm.container.byLifecycleSupport[CanBeAccessed] {

		accessible := component.Instance.(Accessible)
		if err := accessible.AllowAccess(); err != nil {
			return err
		}

	}

	return nil
}

func (lm *LifecycleManager) waitForBlockers(retestInterval time.Duration, maxTries int, warnAfterTries int) error {

	var names []string

	for i := 0; i < maxTries; i++ {

		notReady, cNames := lm.countBlocking(i > warnAfterTries)
		names = cNames

		if notReady == 0 {
			return nil
		} else {
			time.Sleep(retestInterval)
		}
	}

	message := fmt.Sprintf("Startup blocked by %v", names)

	return errors.New(message)

}

func (lm *LifecycleManager) StopAll() error {

	return lm.StopComponents(lm.container.byLifecycleSupport[CanStop])

}

func (lm *LifecycleManager) StopComponents(comps []*Component) error {

	for _, s := range comps {

		s.Instance.(Stoppable).PrepareToStop()
	}

	lm.waitForReadyToStop(5*time.Second, 10, 3)

	for _, s := range comps {

		err := s.Instance.(Stoppable).Stop()

		if err != nil {
			lm.FrameworkLogger.LogErrorf("%s did not stop cleanly %s", s.Name, err)
		}

	}

	return nil
}

func (lm *LifecycleManager) waitForReadyToStop(retestInterval time.Duration, maxTries int, warnAfterTries int) {

	for i := 0; i < maxTries; i++ {

		notReady := lm.countNotReady(i > warnAfterTries)

		if notReady == 0 {
			return
		} else {
			time.Sleep(retestInterval)
		}
	}

	lm.FrameworkLogger.LogFatalf("Some components not ready to stop, stopping anyway")

}

func (lm *LifecycleManager) countBlocking(warn bool) (int, []string) {

	notReady := 0
	names := []string{}

	for _, c := range lm.container.byLifecycleSupport[CanBlockStart] {
		ab := c.Instance.(AccessibilityBlocker)

		block, err := ab.BlockAccess()

		if block {
			notReady += 1
			names = append(names, c.Name)
			if warn {
				if err != nil {
					lm.FrameworkLogger.LogErrorf("%s blocking startup: %s", c.Name, err)
				} else {
					lm.FrameworkLogger.LogErrorf("%s blocking startup (no reason given)", c.Name)
				}

			}
		}

	}

	return notReady, names
}

func (lm *LifecycleManager) countNotReady(warn bool) int {

	notReady := 0

	for _, c := range lm.container.byLifecycleSupport[CanStop] {
		s := c.Instance.(Stoppable)

		ready, err := s.ReadyToStop()

		if !ready {
			notReady += 1

			if warn {
				if err != nil {
					lm.FrameworkLogger.LogWarnf("%s is not ready to stop: %s", c.Name, err)
				} else {
					lm.FrameworkLogger.LogWarnf("%s is not ready to stop (no reason given)", c.Name)
				}

			}
		}

	}

	return notReady
}
