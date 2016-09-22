package instance

import (
	"os"
	"time"
)

const FrameworkPrefix = "grnc"

func ExitError() {
	os.Exit(1)
}

func ExitNormal() {
	os.Exit(0)
}

type System struct {
	BlockIntervalMS      time.Duration //The interval (in milliseconds) between checks of blocking components during the start->available lifecycle
	BlockRetries         int           //How many chances blocking components should be given to declare themselves ready before the application exits with an error
	BlockTriesBeforeWarn int           //How many times a blocking component can declare it is blocked before a warning message is logged
	FlushMergedConfig    bool          //If the merged JSON configuration files should be discarded after startup is complete.
	GCAfterConfigure     bool          //If a garbage collection should be invoked after the container has finished populating and configuring the IoC container.
	GCAfterStart         bool          //If a garbage collection should be invoked after the container has called StartComponent on all components (but before AllowAccess).
	StopIntervalMS       time.Duration //The interval (in milliseconds) between checks of stoppable components to see if they are ready to be stopped.
	StopRetries          int           //How many chances stoppable components should be given to declare themselves ready to stop before the container stops them anyway.
	StopTriesBeforeWarn  int           //How many times a stoppable component can declare it is not ready to stop before a warning message is logged

}

const InstanceIDComponent = FrameworkPrefix + "InstanceIdentifier"

type InstanceIdentifier struct {
	ID string
}

type InstanceIdentifierReceiver interface {
	RegisterInstanceID(*InstanceIdentifier)
}
