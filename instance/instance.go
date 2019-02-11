// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package instance defines functionality for identifying and tuning an individual instance of a Granitic application.

It is anticipated that most Granitic applications will be deployed as part of a group behind a load-balancer or some other
service discovery mechanism. In these circumstances, it is important that each individual instance can be named. So other
services can identify it.

This package defines a type for storing an instance ID and an interface to be implemented
by any component that needs to be aware of the ID of the current instance. An instance ID is specified when starting
Granitic via a command line argument or an InitialSettings struct. See the documentation for the granitic pacakge for
more detail.

This package also defines the data structure for the implicit System facility (which cannot be disabled, but can be configured).
The System facility controls start up and shutdown behaviour as well as some memory management settings. See
http://granitic.io/1.0/ref/system
*/
package instance

import (
	"os"
	"time"
)

// The prefix for all Granitic owned components stored in the IoC container to namespace them apart from user defined
// components. It is strongly recommended (but not enforced) that user components do not use this prefix for their own names.
const FrameworkPrefix = "grnc"

// Immediately exit the application with the return value 1 to signify a problem. Does not cleanly stop IoC components.
func ExitError() {
	os.Exit(1)
}

// Immediately exit the application with the return value 0 to signify a normal exit. Does not cleanly stop IoC components.
func ExitNormal() {
	os.Exit(0)
}

// A structure used by the implicit System facility to control start, stop and some memory management behaviour.
type System struct {
	//The interval (in milliseconds) between checks of blocking components during the start->available lifecycle phase transition.
	BlockIntervalMS time.Duration

	//How many chances blocking components should be given to declare themselves ready before the application exits with an error.
	BlockRetries int

	//How many times a blocking component can declare it is blocked before a warning message is logged.
	BlockTriesBeforeWarn int

	//If the merged JSON configuration files should be discarded after startup is complete.
	FlushMergedConfig bool

	//If a garbage collection should be invoked after the container has finished populating and configuring the IoC container.
	GCAfterConfigure bool

	//If a garbage collection should be invoked after the container has called StartComponent on all components (but before AllowAccess).
	GCAfterStart bool

	//The interval (in milliseconds) between checks of stoppable components to see if they are ready to be stopped.
	StopIntervalMS time.Duration

	//How many chances stoppable components should be given to declare themselves ready to stop before the container stops them anyway.
	StopRetries int

	//How many times a stoppable component can declare it is not ready to stop before a warning message is logged
	StopTriesBeforeWarn int
}

// The name of the component in the IoC container holding an instance ID.
const IDComponent = FrameworkPrefix + "InstanceIdentifier"

// A structure used to store the ID of a particular instance of a Granitic application. See the granitic package
// documentation for instructions on how to define the ID at application start time.
type Identifier struct {
	// A identifier for this instance application.
	ID string
}

// Implemented by any component that needs to be aware of the ID of the current application instance.
type Receiver interface {
	// RegisterInstanceID is automatically called by the IoC container to inject the instance ID.
	RegisterInstanceID(*Identifier)
}
