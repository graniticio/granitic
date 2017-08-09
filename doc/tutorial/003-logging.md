# Tutorial - Logging

## What you'll learn

 1. How Granitic handles logging from your code and from Grantic's built-in components
 2. How to adjust the amount of logging shown
 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 2](002-configuration.md) or open a terminal and run

<pre>
go get github.com/graniticio/granitic-examples
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 3
</pre>

## Related GoDoc

https://godoc.org/github.com/graniticio/granitic/logging

## Logging

Logging in Grantic is designed to allow developers to have fine-grained control over which components output logging information. There are two main
concepts to be aware of:

 * Loggers - the components that format log messages and choose whether or not to display them based on the severity assigned to the message.
 * Severity - the seriousness of a message to be log.
 
Severities are (in ascending order)  <code>TRACE, DEBUG, INFO, WARN, ERROR, FATAL</code>. See the [GoDoc for more detail](https://godoc.org/github.com/graniticio/granitic/logging)

In order for your code to log messages, it will need to have access to a Logger. Granitic has two built-in Loggers - the <code>ApplicationLogger</code> and the <code>FrameworkLogger</code>. 
As the names suggest, the <code>FrameworkLogger</code> is intended for internal Granitic components and the <code>ApplicationLogger</code> is for your application's code.

As almost all every component you build will probably need the <code>ApplicationLogger</code>, Granitic has a built-in [ComponentDecorator](https://godoc.org/github.com/graniticio/granitic/ioc#ComponentDecorator) that automatically
injects a reference to the <code>ApplicationLogger</code> or any component with a member variable:

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
cd $GOPATH/grnc-tutorial/recordstore
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

##Global log level

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

The global log level means that any message with a severity equal to or greater than the specified

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
    }
}
```

If you stop and restart your recordstore application and refresh or re-visit [http://localhost:8080/artist](http://localhost:8080/artist),
you'll see that the INFO message is no longer displayed.

Notice that you are still seeing other INFO messages from components whose names start <code>grnc</code>. These components
are built-in Granitic components so use the <code>FrameworkLogger</code> which has its own <code>GlobalLogLevel</code>

##Component specific log levels

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


