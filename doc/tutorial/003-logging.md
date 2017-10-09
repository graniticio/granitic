# Tutorial - Logging

## What you'll learn

 1. How Granitic handles logging from your code.
 2. How to adjust which messages are logged.
 3. How to change logging behaviour at runtime.
 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 2](002-configuration.md) or open a terminal and run

<pre>
go get github.com/graniticio/granitic-examples
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 3
</pre>

## Logging

Logging in Granitic gives developers fine-grained control over which components output logging information. There are two main
concepts for you to become familiar with:

 * Loggers - the components that format log messages and choose whether or not to write them to a console or file them based on the severity assigned to the message.
 * Severity - the importance of a message to be logged.
 
Severities are (in ascending order of importance)  <code>TRACE, DEBUG, INFO, WARN, ERROR, FATAL</code>. See the [GoDoc for more detail](https://godoc.org/github.com/graniticio/granitic/logging)

Your code will dispatch log messages through a Granitic component called a <code>Logger</code>. Granitic has two built-in Loggers - the <code>ApplicationLogger</code> and the <code>FrameworkLogger</code>. 
As the names suggest, the <code>FrameworkLogger</code> is used by internal Granitic components and the <code>ApplicationLogger</code> is for your application's code.

As the majority of components that you build will need access to the <code>ApplicationLogger</code>, Granitic has a built-in [ComponentDecorator](https://godoc.org/github.com/graniticio/granitic/ioc#ComponentDecorator) that automatically
injects a reference to the <code>ApplicationLogger</code> into any of your components with a member variable that is exactly:

```go
    Log logging.Logger
```

To see this in action, modify the file <code>endpoint/artist.go</code> so it looks like:
                              
```go
package endpoint

import (
	"context"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

type ArtistLogic struct {
	EnvLabel string
	Log      logging.Logger
}

func (al *ArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

	a := new(ArtistDetail)
	a.Name = "Hello, World from " + al.EnvLabel

	res.Body = a

	l := al.Log
	l.LogInfof("Environment is set to '%s'", al.EnvLabel)
	l.LogTracef("Request served")

}

type ArtistDetail struct {
	Name string
}
```

and run

<pre>
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind
go build
./recordstore -c resource/config,resource/env/production.json
</pre>

Keep the terminal window visible and visit [http://localhost:8080/artist](http://localhost:8080/artist) and you'll see a line appear in your terminal similar to:

<pre>
09/Aug/2017:14:50:24 Z INFO  artistLogic Environment is set to 'PROD'
</pre>

This line shows:

 * The timestamp of when the message was logged
 * The severity of the message (INFO)
 * The name of the component issuing the message (artistLogic)
 * The message itself

### Unit tests

You may find the function <code>logging.CreateAnonymousLogger</code> useful when writing unit tests for code that relies
on having access to a <code>Logger</code>

## Global log level

You may have noticed that the <code>INFO</code> level message is shown but the <code>TRACE</code> message is not.This is because the global log level for the <code>ApplicationLogger</code> 
is also set to INFO in the facility configuration file <pre>$GRANITIC_HOME/resource/facility/config/logging.json</pre> which looks something like:

```json
{
  "LogWriting": {
    "EnableConsoleLogging": true,
    "EnableFileLogging": false,
    "File": {
      "LogPath": "./granitic.log",
      "BufferSize": 50
    },
    "Format": {
      "UtcTimes":     true,
      "Unset": "-"
    }
  },

  "FrameworkLogger":{
    "GlobalLogLevel": "INFO"
  },
  "ApplicationLogger":{
    "GlobalLogLevel": "INFO"
  }
}
```

The global log level means that only messages with a severity equal to or greater than the global log level with be logged.

You can override the global log level for you own application. Open <pre>resource/config/config.json</pre> and edit it
so it looks like:

```json
{
    "Facilities": {
            "HttpServer": true,
            "JsonWs": true
    },
    
    "ApplicationLogger":{
            "GlobalLogLevel": "ERROR"
    },
    
    "environment": {
            "label": "UNSET"
    }
}
```

If you stop and restart your recordstore application and refresh or re-visit [http://localhost:8080/artist](http://localhost:8080/artist),
you'll see that the INFO message is no longer displayed.

Notice that you are still seeing other INFO messages from components whose names start <code>grnc</code>. These components
are built-in Granitic components so use the <code>FrameworkLogger</code> which has its own <code>GlobalLogLevel</code>

### File logging

By default Granitic only logs messages to the console. Look at <pre>$GRANITIC_HOME/resource/facility/config/logging.json</pre> to
see how you can enable logging to a file.


## Component specific log levels

Sometimes you want to allow a single component to log messages that are below the global log level (to aid debugging) or to
squelch a component that is too noisy. This can be achieved by setting a log level for a specific component. Modify your
<code>config.json</code> file so that the <code>ApplicationLogger</code> section looks like:

```json
"ApplicationLogger": {
    "GlobalLogLevel": "ERROR",
    "ComponentLogLevels": {
        "artistLogic": "TRACE"
    }
}
```

If you stop and restart your recordstore application and refresh or re-visit [http://localhost:8080/artist](http://localhost:8080/artist),
you'll see an additional message displayed:

<pre>
09/Aug/2017:21:05:55 Z INFO  artistLogic Environment is set to 'PROD'
09/Aug/2017:21:05:55 Z TRACE artistLogic Request served
</pre>

Try setting the <code>artistLogic</code> log level to <code>FATAL</code> to see what happens.

### Runtime control of logging

When investigating problems with production code it can be very helpful to enable lower-priority messages without having
to restart or re-deploy an application. Granitic supports this through the <code>RuntimeCtl</code> facility.

Stop your instance of <code>recordstore</code> and change <pre>resource/config/config.json</pre> so that the <code>Facilities</code>
section looks like:

```json
"Facilities": {
    "HttpServer": true,
    "JsonWs": true,
    "RuntimeCtl": true
}
```

Restart <code>recordstore</code> and you will see a new line in the startup logs:

<pre>10/Aug/2017:06:55:28 Z INFO  grncCtlServer Listening on 9099</pre>

You can now use the [grnc-ctrl command line tool](https://godoc.org/github.com/graniticio/granitic/cmd/grnc-ctl) to issue
commands to <code>recordstore</code> while it is running.

Open a terminal and run:

<pre>grnc-ctl help</pre>

To get a list of the high level actions you can perform with this tool and then:

<pre>
grnc-ctl help global-level
grnc-ctl help log-level
</pre> 

for more information on the commands related to logging. 

Try running:

<pre>grnc-ctl log-level artistLogic FATAL</pre> 

to raise the logging threshold for the <code>artistLogic</code> component to <code>FATAL</code>

Any changes you make with the <code>grnc-ctl</code> tool are non-permanent and will be reset the next time you start
your application.

## Recap

 * Granitic can inject a <code>Logger</code> into your application code.
 * You can log at different levels of severity.
 * You can set the global severity level at which messages will be logged in configuration.
 * You can override this global level for individual components.
 * You can change both the global and component-specific levels at runtime using <code>grnc-ctl</code>
 
## Further reading

 * [Logging GoDoc](https://godoc.org/github.com/graniticio/granitic/logging)
 * [grnc-ctrl usage](https://godoc.org/github.com/graniticio/granitic/cmd/grnc-ctl)
 
 
## Next

The next tutorial covers the [capture of data from web-service calls](004-data-capture.md)  


