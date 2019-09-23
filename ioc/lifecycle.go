// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ioc

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
	"os"
	"runtime"
	"time"
)

// LifecycleSupport is an enumeration able used to categorise types by the lifecycle events they can react to.
type LifecycleSupport int

const (
	// None indicates that the component doesn't support any lifecycle events
	None = iota
	// CanStart indicates that the component is Startable
	CanStart
	// CanStop indicates that the component is Stoppable
	CanStop
	// CanSuspend indicates that the component is Suspendable
	CanSuspend
	// CanBlockStart indicates that the component can block the start process if it is not ready
	CanBlockStart
	// CanBeAccessed indicates that the component is Accessible
	CanBeAccessed
)

/*
Startable is implemented by components that need to perform some initialisation before they are ready to run or use.

Components that provide services outside of the application (like HTTP servers or queue listeners) should consider implementing
Accessible in addition to Startable.
*/
type Startable interface {
	// StartComponent performs initialisation and may start listeners/servers.
	StartComponent() error
}

/*
Suspendable is implemented by components that are able to temporarily halt and then resume their activity at some later point in time.
*/
type Suspendable interface {
	// Suspend causes the component to stop performing its primary function until Resume is called.
	Suspend() error

	// Resume causes the component to resume its primary function.
	Resume() error
}

/*
Stoppable is implemented by components that need to be given the opportunity to release resources or perform shutdown activities
before an application is halted.

See https://granitic.io/ref/system-configuration for information about the number of times ReadyToStop is called, and how the
interval between these calls, can be adjusted for your application.
*/
type Stoppable interface {
	// PrepareToStop gives notice to the component that the application is about to halt. The implementation of this method
	// should cause the component to block any new requests for work.
	PrepareToStop()

	// ReadyToStop is called by the container to query whether or not the component is ready to shutdown. A component might
	// return false if it is still processing or running a job. If the component returns false, it may optionally return
	// an error to explain why it is not ready.
	ReadyToStop() (bool, error)

	// Stop is an instruction by the container to immediately terminate any running processes and release any resources.
	// If the component is unable to do so, it may return an error, but the application will still stop.
	Stop() error
}

/*
AccessibilityBlocker is implemented by components that MUST be ready before an application can be made accessible. For example, a component connecting
to a critical external system might implement AccessibilityBlocker to prevent an HTTP server making an API available until
a connection to the critical system is established.

See https://granitic.io/ref/system-configuration for information about the number of times BlockAccess is called, and how the
interval between these calls can be adjusted for your application.
*/
type AccessibilityBlocker interface {

	// BlockAccess returns true if the component wants to prevent the application from becoming ready and accessible. An
	// optional error message can be returned explaining why.
	BlockAccess() (bool, error)
}

/*
Accessible is implemented by components that require a final phase of initialisation to make themselves outside of the application.
Typically implemented by HTTP servers and message queue listeners to start listening on TCP ports.
*/
type Accessible interface {
	// AllowAccess is called by the container as the final stage of making an application ready.
	AllowAccess() error
}

/*
LifecycleManager provides an interface to the components to allow lifecycle methods/events to be applied to all
components or a subset of the components.
*/
type LifecycleManager struct {
	container       *ComponentContainer
	FrameworkLogger logging.Logger
	system          *instance.System
}

// StartAll finds all Startable and Accessible components runs the Start/Block/Accessible cycle.
func (lm *LifecycleManager) StartAll() error {

	defer func() {
		if r := recover(); r != nil {
			lm.FrameworkLogger.LogErrorfWithTrace("Panic recovered while starting components components %s", r)
			os.Exit(-1)
		}
	}()

	startable := lm.container.byLifecycleSupport[CanStart]
	accessible := lm.container.byLifecycleSupport[CanBeAccessed]

	return lm.start(startable, accessible)
}

/*
Start starts the supplied components, waits for any access-blocking components to be ready, then makes all
components accessible. See GoDoc for Startable, AccessibilityBlocker and Accessible for more details.
*/
func (lm *LifecycleManager) Start(startable []*Component) error {

	accessible := make([]*Component, 0)

	for _, c := range startable {

		if _, found := c.Instance.(Accessible); found {
			accessible = append(accessible, c)
		}

	}

	return lm.start(startable, accessible)

}

func (lm *LifecycleManager) start(start []*Component, access []*Component) error {

	for _, component := range start {

		startable := component.Instance.(Startable)

		if err := startable.StartComponent(); err != nil {
			message := fmt.Sprintf("Unable to start %s: %s", component.Name, err)
			return errors.New(message)
		}

	}

	if lm.system.GCAfterStart {
		runtime.GC()
	}

	if len(lm.container.byLifecycleSupport[CanBlockStart]) != 0 {

		sys := lm.system
		bi := sys.BlockIntervalMS * time.Millisecond

		if err := lm.waitForBlockers(bi, sys.BlockRetries, sys.BlockTriesBeforeWarn); err != nil {
			return err
		}

	}

	for _, component := range access {

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

		if notReady != 0 {
			time.Sleep(retestInterval)

		} else {
			return nil
		}
	}

	message := fmt.Sprintf("Startup blocked by %v", names)

	return errors.New(message)

}

// StopAll finds all components implementing Stoppable and passes them to Stop
func (lm *LifecycleManager) StopAll() error {

	return lm.StopComponents(lm.container.byLifecycleSupport[CanStop])

}

// SuspendComponents invokes Suspend on all of the supplied components that implement Suspendable
func (lm *LifecycleManager) SuspendComponents(comps []*Component) error {

	for _, c := range comps {

		if err := c.Instance.(Suspendable).Suspend(); err != nil {
			lm.FrameworkLogger.LogErrorf("Problem suspending %s: %s", c.Name, err.Error())
		}

	}

	return nil
}

// ResumeComponents invokes Resume on all of the supplied components that implement Suspendable
func (lm *LifecycleManager) ResumeComponents(comps []*Component) error {

	for _, c := range comps {

		if err := c.Instance.(Suspendable).Resume(); err != nil {
			lm.FrameworkLogger.LogErrorf("Problem resuming %s: %s", c.Name, err.Error())
		}

	}

	return nil
}

/*
StopComponents invokes PrepareToStop on all components then waits for them to be ready to stop by
calling ReadyToStop on each component. If one or more components are not ready, they are given x chances to become
ready with y milliseconds between each check. See https://granitic.io/ref/system-configuration

If all components are ready, or if x has been exceeded, Stop is called on all components.
*/
func (lm *LifecycleManager) StopComponents(comps []*Component) error {

	for _, s := range comps {

		s.Instance.(Stoppable).PrepareToStop()
	}
	sys := lm.system
	si := sys.StopIntervalMS * time.Millisecond

	lm.waitForReadyToStop(si, sys.StopRetries, sys.StopTriesBeforeWarn)

	for _, s := range comps {

		if err := s.Instance.(Stoppable).Stop(); err != nil {
			lm.FrameworkLogger.LogErrorf("%s did not stop cleanly %s", s.Name, err.Error())
		}

	}

	return nil
}

func (lm *LifecycleManager) waitForReadyToStop(retestInterval time.Duration, maxTries int, warnAfterTries int) {

	for i := 0; i < maxTries; i++ {

		notReady := lm.countNotReady(i > warnAfterTries)

		if notReady != 0 {
			time.Sleep(retestInterval)
		} else {
			return
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
			notReady++
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
			notReady++

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
