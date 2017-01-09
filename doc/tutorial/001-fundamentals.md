# Tutorial - Fundamentals

## What you'll learn

 1. How to create a basic Granitic application
 2. Some of Granitic's fundamental principles
 3. How to create a simple web service
 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Have access to a text editor or IDE for writing Go and JSON files and a terminal emulator or command prompt 
 (referred to as a terminal from now on)

## Introduction

This tutorial will show you how to create a simple Granitic application that exposes a JSON web service over HTTP. It's 
a little more in-depth than the usual _'look how much you can achieve in two lines with this powerful framework'_ tutorial, 
as it's a good opportunity to introduce some of Granitic's fundamentals.
 
## Creating a Grantic project

A Granitic project is made up of three sets of assets:

 1. Go source files
 2. JSON application configuration files
 3. JSON component definition files
 
Each of these types of asset will be explained over the course of this tutorial. Granitic includes a command line tool
to create a skeleton project that can be compiled and started.

Run the following in a terminal:

<pre>
cd $GOPATH/src
grnc-project grnc-tutorial grnc-tutorial
</pre>

This will create the following files under $GOPATH/src:

<pre>
/grnc-tutorial
    service.go
    /resource
        /components
            components.json
        /config
            config.json
</pre>

The first argument to <code>grnc-project</code> is a name for your project. 
The second is the location of your project relative to <code>$GOHOME/src</code> - specifying this allows the tool to generate source files
that are ready to use with <code>go build</code> and <code>go install</code> (normally these two arguments would be different to each other as 
most Go projects are not created in the root of <code>$GOHOME/src</code>).

## Starting and stopping your application

This minimal configuration is actually a working Granitic application that can be started and stopped - it just doesn't 
do anything interesting.

Start the application by returning to your terminal and running

<pre>
cd grnc-tutorial
grnc-bind
go build
./grnc-tutorial
</pre>

You should see output similar to:

<pre>

04/Jan/2017:14:41:20 Z INFO  grncInit Starting components
04/Jan/2017:14:41:20 Z INFO  grncInit Ready (startup time 3.866393ms)

</pre>

This means your application has started and is waiting. You can stop it with <code>CTRL+C</code> and will see output similar to

<pre>
04/Jan/2017:14:43:11 Z INFO  grncInit Shutting down (system signal)
</pre>

## Facilities

A <code>facility</code> is Granitic's name for a high-level feature that your application can enable or disable. By default,
most of the features are disabled. You can see which features that are available to you and whether or not they're enabled 
by inspecting the file <code>$GRANITIC_HOME/resource/facility-config/facilities.json</code>:

```javascript
{
  "Facilities": {
    "HttpServer": false,
    "JsonWs": false,
    "XmlWs": false,
    "FrameworkLogging": true,
    "ApplicationLogging": true,
    "QueryManager": false,
    "RdbmsAccess": false,
    "ServiceErrorManager": false,
    "RuntimeCtl": false
  }
}
```

In order to build a JSON web service, you will need to enable two facilities: <code>HttpServer</code> and <code>JsonWs</code> (JSON Web Services).

To do this, open the file <code>/grnc-tutorial/resource/config/config.json</code> and change it so it looks like:

```javascript
{
	"Facilities": {
		"HttpServer": true,
    	"JsonWs": true
	}
}
```
(from now on this file will be referred to as your application's config file)

If you return to your terminal and run:

<pre>
grnc-bind
go build
./grnc-tutorial
</pre>

You'll see an additional line of logging on startup similar to:

<pre>
04/Jan/2017:16:34:27 Z INFO  grncHttpServer Listening on 8080
</pre>

Which shows that a HTTP server is listening on the default port of 8080. Stop the runnning service with <code>CTRL+C</code>

## Adding an endpoint

An <code>endpoint</code> is Granitic's preferred name for code that handles a web service request to a particular URI pattern for a 
particular HTTP method (GET, POST etc). Most of the mechanics of routing a request to your code and converting between
JSON and Go is handled by Granitic, you will be concerned mainly with defining your _endpoint logic_.

Endpoint logic is code in a Go struct implementing the <code>ws.WsRequestProcessor</code> interface.

These tutorials are based on the <code>granitic-examples/recordstore</code> demo application, so let's recreate one of the endpoints
from that application. 

Create the file <code>grnc-tutorial/endpoint/artist.go</code> and set the contents to:

```go
package endpoint

import (
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
)

type ArtistLogic struct {
}

func (al *ArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

	a := new(ArtistDetail)
	a.Name = "Hello, World!"

	res.Body = a
}

type ArtistDetail struct {
	Name string
}
```

This code defines an object implementing the <code>ws.WsRequestProcessor</code> interface and another object that will 
be used to store the results of the web service call, in this case a recording artist with the unlikely name "Hello, World!"


## Turning your code into a component

At the core of Granitic is an Inversion of Control or IoC container. Granitic looks after the lifecycle (creating and destroying)
of the Go objects you define, but needs to be told which objects should be included in your application and how they should 
be configured These definitions are stored in a JSON component definition file which, by default, are stored under 
<code>resource/component</code>

A component is a named instance of a Go object, managed by the IoC container.


Open the file <code>grnc-tutorial/resource/components/components.json</code> and set the content to:

```javascript
{
	"packages": [
		"github.com/graniticio/granitic/ws/handler",
		"grnc-tutorial/endpoint"
	],

	"components": {
		"artistLogic": {
			"type": "endpoint.ArtistLogic"
		},

		"artistHandler": {
			"type": "handler.WsHandler",
			"HTTPMethod": "GET",
			"Logic": "ref:artistLogic",
			"PathPattern": "^/artist"
		}
	}
}
```

A component definition file has two sections. The <code>packages</code> section declares the Go packages containing the 
code that you intend to use as components. The <code>components</code> section declares uniquely named components that 
you want to be managed by Granitic.

The first component, named <code>artistLogic</code> has a <code>type</code> field that specifies that the component should 
be an instance of the <code>ArtistLogic</code> Go struct you wrote above. 

The second declared component, <code>artistHandler</code>, is an instance of <code>ws.WsHandler</code>, a built-in Granitic type.
A <code>ws.WsHandler</code> coordinates the bulk of the request processing lifecycle as well as managing error-handling 
for a web service request.

The minimal configuration in this example specifies the HTTP method that the handler will respond to (GET) 
and a regex for matching against incoming request paths.

The line 

<pre>"Logic": "ref:artistLogic"</pre> 

is an example of a component reference. When <code>artistHandler</code> is instantiated by the Granitic container, the previously declared 
<code>artistLogic</code> component will be used as value for the <code>Logic</code> field on the <code>artistHandler</code> component.

## Binding

Granitic is written in Go because of Go's positioning between C's performance and memory consumption and the 
relative ease-of-use of JVM and CLR languages (Java, C# etc). One of the compromises we accept in using Go is that it 
is a statically-linked language with no facility for creating objects from a text name at runtime - if you want to use 
a type in your application, it must be referenced at compile time.

In order to reconcile the configuration-based approach favoured by Granitic with this limitation, a tool is used to 
generate Go source code from component definition files.

Return to your terminal and run 

<pre>grnc-bind</pre> 

You will notice that a Go source file <code>bindings/bindings.go</code> has been created. You will not 
(and in fact should not) edit this file directly, but feel free to examine it to see what is happening.

*You will need to re-run <code>grnc-bind</code> whenever you change your component definition file*


## Building and testing your application

Every Go application requires an entry point <code>main</code> method. For a Go application that was created using the
<code>grnc-project</code> tool, the <code>main</code> method is in the <code>service.go</code> file at the root of the 
project. For this tutorial, this file will look like:

```go
package main

import "github.com/graniticio/granitic"
import "grnc-tutorial/bindings"

func main() {
	granitic.StartGranitic(bindings.Components())
}
```

This simply takes the objects generated by <code>grnc-bind</code> and passes them to Granitic. For the vast majority of
Granitic applications you will not need to modify or even look at this file.

Return to your terminal and run:

<pre>
go build
./grnc-tutorial
</pre>

Now open a browser and visit:

[http://localhost:8080/artist](http://localhost:8080/artist) or [http://[::1]:8080/artist](http://[::1]:8080/artist)

and you should see the response:

```javascript
{
    "Response":{
        "Name": "Hello, World!"
    }
}
```

## Recap

 * Granitic applications contain Go source files, configuration files and component definition files
 * The <code>grnc-project</code> tool can create an empty, working Granitic application
 * Components are a named instance of a Go object, managed by Granitic's IoC container
 * The <code>grnc-bind</code> tool converts your component definition files into Go source - run the tool whenever
  you change your component definitions.
 
## Next

The next tutorial covers [configuration](002-configuration.md) in more depth
 