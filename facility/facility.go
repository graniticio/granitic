// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package facility defines the high-level features that Granitic makes available to applications.

A facility is Granitic's term for a group of components that together provide a high-level feature to application developers,
like logging or service error management. This package contains several sub-packages, one for each facility that can be
enabled and configured by user applications.

A full description of how facilities can be enabled and configured can be found at https://granitic.io/ref/facilities but a basic description
of how they work follows:

# Enabling and disabling facilities

The features that are available to applications, and whether they are enabled by default or not, are enumerated in the file:

	$GRANITIC_HOME/resource/facility-config/facilities.json

which will look something like:

	{
	  "Facilities": {
		"HTTPServer": false,
		"JSONWs": false,
		"XMLWs": false,
		"FrameworkLogging": true,
		"ApplicationLogging": true,
		"QueryManager": false,
		"RdbmsAccess": false,
		"ServiceErrorManager": false,
		"RuntimeCtl": false,
		"TaskScheduler": false
	  }
	}

This shows that the ApplicationLogging and FrameworkLogging facilities are enabled by default, but everything else is turned
off. If you wanted to enable the HTTPServer facility, you'd add the following to any of your application's configuration files:

	{
	  "Facilities": {
		"HTTPServer": true
	  }
	}

# Configuring facilities

Each facility has a number of default settings that can be found in the file:

	$GRANITIC_HOME/resource/facility-config/facilityname.json

For example, the default configuration for the HTTPServer facility will look something like:

	  {
	    "HTTPServer":{
	      "Port": 8080,
		  "AccessLogging": false,
		  "TooBusyStatus": 503,
		  "AutoFindHandlers": true
	    }
	  }

Any of these settings can be changed by overriding one or more of the fields in your application's configuration file. For example, to
change the port on which your application's HTTP server listens on, you could add the following to any of your application's configuration files:

	{
	  "HTTPServer":{
	    "Port": 9000
	  }
	}
*/
package facility

import (
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
)

// A Builder is responsible for programmatically constructing the objects required to support a facility and,
// where required, adding them to the IoC container.
type Builder interface {
	//BuildAndRegister constructs the components that together constitute the facility and stores them in the IoC
	// container.
	BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error

	//FacilityName returns the facility's unique name. Used to check whether the facility is enabled in configuration.
	FacilityName() string

	//DependsOnFacilities returns the names of other facilities that must be enabled in order for this facility to run
	//correctly.
	DependsOnFacilities() []string
}
