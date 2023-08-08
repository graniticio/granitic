# Tutorial - Capturing data from service calls

## What you'll learn

 1. How Granitic captures data from HTTP web service calls and binds it to your application's data structures
 
## Prerequisites

 1. Follow the Granitic [installation instructions](../installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 3](003-logging.md) or clone the
  [tutorial repo](https://github.com/graniticio/tutorial) and navigate to `json/004/recordstore` in your terminal.

## Data capture

Granitic's core purpose is to streamline the process of building web and micro services in Go. 
One of the most time-consuming aspects of building web services is writing code to read inbound data from an HTTP request 
and binding that data to your application's data structures. This tutorial explains how Granitic makes this easier 
(the next tutorial will cover the [validation](005-validation.md) of the data you capture).

Web services that are provided as HTTP endpoints generally allow callers to supply _data_ in three different parts of 
the HTTP request:

1. The HTTP request's body
2. The [path](https://tools.ietf.org/html/rfc2616#section-3.2.2) component of the URL
3. The [query](https://tools.ietf.org/html/rfc2616#section-3.2.2) component of the URL

or some combination of the above. Granitic automates the process of extracting these three types of data from the 
request and mapping the data to fields on your application's data structures.


### Web service design patterns

Granitic is agnostic as to which web service design pattern (REST, RPC etc.) your application implements. The majority of 
the examples in these tutorials are 'REST-like', but that is not to suggest that Granitic favours REST over other patterns.

### Meta-data

Callers of HTTP web services may also supply _meta-data_ about the request as 
[request headers](https://tools.ietf.org/html/rfc2616#section-5.3). 
Granitic gives you application access to these headers but does not currently offer a way of automatically binding 
those headers to your data structures.


## Path and query binding

Our tutorial application already supports a single GET endpoint to allow us to recover the details of a recording artist. 
Start your tutorial application by opening a terminal, navigating to the folder containing the tutorial application
and running:

<pre>
grnc-bind && go build && ./recordstore
</pre>

and visit [http://localhost:8080/artist](http://localhost:8080/artist) to see what happens when you execute a GET request.

We want to allow the caller of this endpoint to specify _which_ artist they'd like details for. Open `artist/get.go` and
add the following struct to the end of the file.

```go
type ArtistQuery struct {
  ID int
  NormaliseName *types.NilableBool
}
```

Make sure you add `github.com/graniticio/granitic/v2/types` to that file's list of imports 
if your IDE hasn't already done it for you.

You should also change the signature for the method:

```go
  func (gl *GetLogic) Process(ctx context.Context, req *ws.Request, res *ws.Response)
```

to 

```go
  func (gl *GetLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, q *ArtistQuery)
```

As we now want Granitic to use some of the logical contents of the HTTP request (the payload) as data to populate the new
`ArtistQuery` type.

### Nilable types

A side-effect of Go's system of [zero values](https://golang.org/ref/spec#The_zero_value) for variables is that it can 
make recovering data from web service calls ambiguous. For example, if you accept a boolean via a query parameter and 
the value of the boolean is `false`, how does your code know if it's false because:

 * The caller explicitly set the value to false or
 * The caller didn't supply that parameter at all, so the variable just defaulted to false.
 
Granitic's solution this problem is to provide a set of 'nilable' struct versions of primitive types 
([see the Godoc](https://godoc.org/github.com/graniticio/granitic/v2/types)) that provide additional methods to 
indicate whether the value was explicitly set by the caller or was an automatic zero value.

## Configuring path binding

A common REST-like technique is to allow a caller to specify the ID of the required resource (in this case a recording artist)
into the _path_ of the request. E.g. `/artist/1234`. We will configure Granitic to extract that ID and inject it
into the ID field of the `ArtistRequest` struct you defined above - this process is known as _path binding_ in Granitic.

If you open the file `comp-def/common.json` you will see:

```json
"artistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "Logic": {
    "type": "artist.GetLogic",
    "EnvLabel": "$environment.label(DEV)"
  },
  "PathPattern": "^/artist"
}
```

The component `artistHandler` is an instance of [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsHandler) 
It provides all of the automatic web-service processing features supported by Granitic.

We can define how path binding will work by configuring this component. Change the definition of your `artistHandler` component so it looks like:


```json
"artistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "Logic": {
    "type": "artist.GetLogic",
    "EnvLabel": "$environment.label(DEV)"
  },
  "PathPattern": "^/artist/([\\d]+)[/]?$",
  "BindPathParams": ["ID"]
}
```

### Parsing data into a target struct

The new method signature:

```go
  func (gl *GetLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, q *ArtistQuery)
```

on your logic component tells Granitic that incoming data (in the HTTP request body, path and parameters) 
should be parsed into a new instance of the struct `artist.ArtistQuery`. In Granitic terminology, this struct is called
the _target_.


Change your `GetLogic` struct to look like:

```go
type GetLogic struct {
  EnvLabel string
  Log      logging.Logger
}

func (gl *GetLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, aq *ArtistQuery) {

	gl.Log.LogTracef("Request for artist with ID %d", aq.ID)

	name := "Some Artist"

	res.Body = Info{
		Name: name,
	}

}
```


### Capturing data in request paths

We've altered the regular expression that this endpoint expects to 

`^/artist/([\d]+)[/]?$`

so that in order to match an incoming request that request's path: 

 1. Must start with the string `/artist/'
 1. Must then have a sequence of one or more digits (the ID of the artist)
 1. May optionally have a trailing `/` character (this is a convenience for the caller)


We've defined a [regular expression group](https://www.regular-expressions.info/refcapture.html) (round brackets) around 
the part of the path that will contain the requested ID.

We've also added a new field `BindPathParams` and set it to an array of strings. The number of strings in this
array should match the number of groups in the `PathPattern` regex. Here we are saying that the value of the first regex group
should be injected into a field called 'ID' on the _target_ for data extraction, which in this case is `artist.ArtistQuery`


### Testing your changes

Stop, rebuild and restart your application:

<pre>
grnc-bind && go build && ./recordstore
</pre>

Visiting [http://localhost:8080/artist](http://localhost:8080/artist) will now result in a 404 Not Found error, 
but visiting [http://localhost:8080/artist/1234](http://localhost:8080/artist/1234) should result in a response and a 
log line similar to:

<pre>
08/Aug/2023:10:28:10 Z TRACE [artistHandlerLogic] Request for artist with ID 1234
</pre>


## Binding query parameters

The technique for binding query parameters (key-value pairs in the request URL after the `?` symbol) 
to your 'target' object is very similar to that used for binding path parameters. 

We will demonstrate this by adding support for a boolean parameter to normalise the name of the artist in the response.

Edit your `comp-def/common.json` file  and _add_ the following to the definition of your `artistHandler` component:

```json
   "FieldQueryParam": {
      "NormaliseName": "normalise"
   }
``` 

This is instructing Granitic to populate a field (`NormaliseName`) on your target with the value of a particular query
parameter (`normalise`). 


Modify your `GetLogic.ProcessPayload` method to look like

```go
func (al *GetLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, aq *ArtistQuery) {

	al.Log.LogTracef("Request for artist with ID %d", aq.ID)

	name := "Some Artist"

	if aq.NormaliseName.Bool() {
		name = strings.ToUpper(name)
	}

	res.Body = Info{
		Name: name,
	}
}
```

Rebuild and restart your application. Visiting [http://localhost:8080/artist/1234?normalise=true](http://localhost:8080/artist/1234?normalise=true) 
will now cause the returned artist's name to be capitalised.
 
## Extracting data from the request body

Path parameters and query parameters are only useful for submitting limited amounts of semi-structured data to a web service. More common is to use a POST or PUT request to include more complex data in the body of an HTTP request. Granitic has built-in support for accepting data in an HTTP request
body as JSON or XML. The following examples all use JSON, refer to the [facility/ws](https://godoc.org/github.com/graniticio/granitic/v2/facility/ws) and
[ws/xml](https://godoc.org/github.com/graniticio/granitic/v2/facility/ws) GoDoc to discover how to use XML instead.

We will create a new endpoint to accept details of a new artist as POSTed JSON. 

Create a new file `artist/post.go` with the following contents:

```go
package artist

import (
  "context"
  "github.com/graniticio/granitic/v2/logging"
  "github.com/graniticio/granitic/v2/types"
  "github.com/graniticio/granitic/v2/ws"
)

type PostLogic struct {
  Log      logging.Logger
}

func (pl *PostLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, s *Submission) {
  pl.Log.LogInfof("New artist: '%s'", s.Name)

  res.Body = CreatedResource{0}
}

type Submission struct {
  Name *types.NilableString
  FirstYearActive *types.NilableInt64
}

type CreatedResource struct{
  ID int64
}
```

This defines new endpoint logic that will expect a populated `Submission` to be supplied
as the [ws.WsRequest](https://godoc.org/github.com/graniticio/granitic/v2/ws#WsRequest).RequestBody. 

In order to have this code invoked, we will need to add the following to the `components` section of our 
`comp-def/common.json` file:

```json
"submitArtistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "POST",
  "PathPattern": "^/artist[/]?$",
  "Logic": {
    "type": "artist.PostLogic"
  }
}
```

Check your `config/base.json` file and make sure that `ApplicationLogger.GlobalLogLevel` is set to `INFO`

<pre>
grnc-bind && go build && ./recordstore
</pre>

### Testing POST services

Testing POST and PUT services is more complex than GET services as browsers don't generally have built-in mechanisms for 
setting the body of a request. There are several browser extensions available that facilitate this sort of testing. The following
instructions are based on [Postman](https://www.postman.com/)

## POST a new  artist


1. Open Postman
2. Press the 'New' button next to 'My Workspace' in the top-left of the screen
3. Choose 'HTTP'
4. Set the HTTP method to `POST` and the URL to `http://localhost:8080/artist`
1. Click on **Body** and click on the `raw` radio button
1. Enter the 'test JSON' below into the large grey text area
1. Press `SEND`
1. You should receive a JSON formatted response with an ID of 0 and see a log line similar to: `08/Aug/2023:11:44:54 Z INFO  [submitArtistHandlerLogic] New artist: 'Another Artist'`

### Test JSON

```json
{
  "Name": "Another Artist",
  "FirstYearActive": 2010
}
```
### Saving this request for later tutorials

In Postman:

1. Click on the dropdown next to the `Save` button and choose `Save As`
2. Set the request name to 'Artist POST'
3. Click on `New Collection` at the bottom of the dialog
4. Name the new collection 'Granitic Tutorials' and press create
5. Press `Save`

## Recap

 * Granitic can extract data from the path, query and body of an HTTP request and bind it to your custom Go structs.
 * All this behaviour is configurable by changing the configuration of your handler components
 * Handler components are instances of [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsHandler)
 
## Further reading

 * [Granitic web service processing GoDoc](https://godoc.org/github.com/graniticio/granitic/v2/ws)
 * [Granitic web service facility GoDoc](https://godoc.org/github.com/graniticio/granitic/v2/facility/ws)
 
 
## Next

The next tutorial covers the [validation of data submitted to web services](005-validation.md)

