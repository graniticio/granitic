# Tutorial - Fundamentals

## What you'll learn

 1. How to create a basic Granitic application
 2. Some of Granitic's fundamental principles
 3. How to create a simple web service
 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/v2/blob/master/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Have access to a text editor or IDE for writing Go and JSON files and a terminal emulator or command prompt 
 (referred to as a terminal from now on)
 
#### Note for Windows users

Refer to the Windows notes in the [before you start](000-before-you-start.md) tutorial 

## Introduction

This tutorial will show you how to create a simple Granitic application that exposes a JSON web service over HTTP. It will
also introduce you to some of Granitic's fundamentals.
 
## Creating a Grantic project

A Granitic project is made up of three sets of assets:

 1. Go source files
 2. JSON application configuration files
 3. JSON component definition files
 
Each of these types of asset will be explained over the course of this tutorial. Granitic includes a command line tool
to create a skeleton project that can be compiled and started.

Run the following in a terminal:

<pre>
cd ~
grnc-project recordstore
</pre>

This will create the following files under your home directory:

<pre>
recordstore
  .gitignore
  main.go
  go mod
  comp-def/
    common.json
  config/
    base.json
</pre>

## Starting and stopping your application

This minimal configuration is actually a working Granitic application that can be started and stopped - it just doesn't 
do anything interesting.

Start the application by returning to your terminal and running

<pre>
cd recordstore
go mod download
grnc-bind
go build
./recordstore
</pre>

You should see output similar to:

<pre>
01/Mar/2019:14:59:52 Z INFO  [grncInit] Starting components
01/Mar/2019:14:59:52 Z INFO  [grncInit] Ready (startup time 1.68654ms)
</pre>

This means your application has started and is waiting. You can stop it with `CTRL+C` and will see output similar to

<pre>
01/Mar/2019:15:00:36 Z INFO  [grncInit] Shutting down (system signal)
</pre>

## Facilities

A `facility` is Granitic's name for a high-level feature that your application can enable or disable. By default,
most of the features are disabled. You can see which features are available to your applications and whether or not they're enabled 
by inspecting the file `facility/config/active.json` in your Granitic installation folder:

```json
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
```

In order to build a JSON web service that will listen for an handler HTTP requests, you will need to enable two facilities: 
`HTTPServer` and `JSONWs` (JSON Web Services).

We do this by <i>overriding</i> the default setting for each facility. To do this, open the JSON `config/base.json` 
that was generated for you and change it so it looks like:

```json
{
  "Facilities": {
    "HTTPServer": true,
    "JSONWs": true
  }
}
```
(from now on this file will be referred to as your base config file)

If you return to your terminal and run:

<pre>
grnc-bind && go build && ./recordstore
</pre>

You'll see an additional line of logging on startup similar to:

<pre>
01/Mar/2019:15:12:29 Z INFO  [grncHTTPServer] Listening on 8080
</pre>

Which shows that an HTTP server is listening for web service requests on the default port of 8080. Stop the running 
service with `CTRL+C`

## Adding an endpoint

An `endpoint` is Granitic's preferred name for code that handles a web service request to a particular URI pattern for a 
particular HTTP method (GET, POST etc). Most of the mechanics of routing a request to your code and converting between
JSON and your custom Go code is handled by Granitic, you will be concerned mainly with defining your _endpoint logic_.

Endpoint logic is code in a Go struct that has a member function that Granitic can call to pass control of a request
once the framework has completed its automatic steps.

Create the file `artist/get.go` in your `recordstore` project and set the contents to:

```go
package artist

import (
  "github.com/graniticio/granitic/v2/ws"
  "context"
)

type GetLogic struct {}

func (gl *GetLogic) Process(ctx context.Context, req *ws.Request, res *ws.Response) {
	
  res.Body = Info{
    Name: "Hello, World!",
  }

}

type Info struct {
  Name string
}
```

This code defines an object implementing the `ws.WsRequestProcessor` interface and another object that will 
be used to store the results of the web service call, in this case a recording artist with the unlikely name "Hello, World!"


## Turning your code into a component

At the core of Granitic is an Inversion of Control (IoC) container, sometimes also called a Dependency Injection framework. 
Granitic looks after the lifecycle (creating and destroying) of the Go objects you define, but needs to be told which 
objects should be included in your application and how they should be configured. 

These definitions are stored in JSON _component definition files_ which, by default, are stored in your project in a folder called `comp-def`. 
You can have as many files as you like in this folder, and it is recommend you group related components in to separate
named files.


Open the file `comp-def/common.json` and set the content to:

```json
{
  "packages": [
    "github.com/graniticio/granitic/v2/ws/handler",
    "recordstore/artist"
  ],

  "components": {
    "artistHandler": {
      "type": "handler.WsHandler",
      "HTTPMethod": "GET",
      "Logic": {
        "type": "artist.GetLogic"  
      },
      "PathPattern": "^/artist"
    }
  }
}
```

A component definition file has two sections. The `packages` section declares the Go packages containing the 
code that you intend to use as components. The `components` section declares uniquely named components that 
you want to be managed by Granitic.

The sole component in this file, `artistHandler`, is an instance of `handler.WsHandler`, a built-in Granitic 
type. A `ws.WsHandler` coordinates the bulk of the request processing lifecycle as well as managing error-handling 
for a web service request.

One of the fields on this component, `Logic`, expects another component to be injected into to it. In this case
we're defining the component that the needs to be injected inline as a _nested component_. Note that the nested component
doesn't need a name, just a type - this case the  `artist.GetLogic` struct we defined above.

The minimal configuration in this example specifies the HTTP method that the handler will respond to (GET) 
and a regex for matching against incoming request paths.

## Binding

Granitic is written in Go because of Go's positioning between C's performance and memory consumption and the 
relative ease-of-use of JVM and CLR languages (Java, C# etc). One of the compromises we accept in using Go is that it 
is a statically-linked language with no facility for creating objects from a text name at runtime - if you want to use 
a type in your application, it must be referenced at compile time.

In order to reconcile the configuration-based approach favoured by Granitic with this limitation, a tool is used to 
generate Go source code from component definition files.

Return to your terminal and run 

<pre>grnc-bind</pre> 

You will notice that a Go source file `bindings/bindings.go` has been created. You will not 
(and in fact should not) edit this file directly, but feel free to examine it to see what is happening.

*You will need to re-run `grnc-bind` whenever you change your component definition file*


## Building and testing your application

Every Go application requires an entry point `main` method. For a Go application that was created using the
`grnc-project` tool, the `main` method is in the `main.go` file at the root of the 
project. For this tutorial, this file will look like:

```go
package main

import "github.com/graniticio/granitic/v2"
import "recordstore/bindings"

func main() {
  granitic.StartGranitic(bindings.Components())
}

```

This simply takes the objects generated by `grnc-bind` and passes them to Granitic. For the vast majority of
Granitic applications you will not need to modify or even look at this file.

Return to your terminal and run:

<pre>
go build && ./recordstore
</pre>

Now open a browser and visit:

[http://localhost:8080/artist](http://localhost:8080/artist) or [http://[::1]:8080/artist](http://[::1]:8080/artist)

and you should see the response:

```json
{
  "Name": "Hello, World!"
}
```

## Recap

 * Granitic applications contain Go source files, configuration files and component definition files
 * The `grnc-project` tool can create an empty, working Granitic application
 * Components are a named instance of a Go object, managed by Granitic's IoC container
 * The `grnc-bind` tool converts your component definition files into Go source - run the tool whenever
  you change your component definitions.
 
## Next

The next tutorial covers [configuration](002-configuration.md) in more depth
 