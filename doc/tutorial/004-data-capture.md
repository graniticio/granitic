# Tutorial - Capturing data from service calls

## What you'll learn

 1. How Granitic captures data from HTTP web service calls and binds it to your application's data structures
 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 3](003-logging.md) or open a terminal and run

<pre>
go get github.com/graniticio/granitic-examples
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 4
</pre>

## Data capture

Granitic's core purpose is to streamline the process of building web and micro services in Go. 
One of the most time consuming aspects of buliding web services is writing code to read inbound data from an HTTP request 
and binding that data to your application's data structures. This tutorial explains how Granitic makes this easier (the next tutorial will cover the [validation](005-validation.md) of the data you capture).

Web services that are provided as HTTP endpoints generally allow callers to supply _data_ in three different parts of the HTTP request:

1. The HTTP request's body
2. The [path](https://tools.ietf.org/html/rfc2616#section-3.2.2) component of the URL
3. The [query](https://tools.ietf.org/html/rfc2616#section-3.2.2) component of the URL

or some combination of the above. 

Granitic provides support for automating the binding 


### Web service design patterns

Granitic is agnostic as to which design pattern (REST, RPC) your application follows. The majority of the examples in these tutorials are 'REST-like',
but that is not to suggest that Granitic favours REST over other patterns.

### Meta-data

Callers of HTTP web services may also supply _meta-data_ about the request as [request headers](https://tools.ietf.org/html/rfc2616#section-5.3). Granitic gives you application access to these headers but does not
currently offer a way of automatically binding those headers to your data structures.


## Path and query binding

Our tutorial application already supports a single GET endpoint to allow us to recover the details of a recording artist. Start your tutorial application:

<pre>
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind
go build
./recordstore -c resource/config,resource/env/production.json
</pre>

and visit [http://localhost:8080/artist](http://localhost:8080/artist) to see what happens when you execute a GET request.

We'll want to allow a called to specify _which_ artist they'd like details for. Open <code>endpoint/artist.go</code> and
add the following struct to the end of the file.

```go
type ArtistRequest struct {
	Id int
	NormaliseName *types.NilableBool
}
```

Make sure you add <code>github.com/graniticio/granitic/types</code> to that file's list of imports if your IDE hasn't already done it for you.

### Nilable types

A side-effect of Go's system of [zero values](https://golang.org/ref/spec#The_zero_value) for variables is that it can make recovering data from web service calls
ambiguous. For example, if you accept a boolean via a query parameter and the value of the boolean is <code>false</code>, how does your code know if it's false because:

 * The caller explictly set the value to false or
 * The caller didn't supply that parameter at all, so the variable just defaulted to false.
 
Granitic's soltuion this problem is to provide a set of 'nilable' struct versions of primitive types ([see the Godoc](https://godoc.org/github.com/graniticio/granitic/types)) that provide additional methods to indicate whether the value was explicity set by the caller
or was an automatic zero value.

## Configuring path binding

A common REST-like technique is to allow a caller to specify the ID of the required resource, in this case a recording artist,
into the _path_ of the request. E.g. <code>/artist/1234</code>. We will configure Granitic to extract that ID and inject it
into the Id field of the <code>ArtistRequest</code> struct you defined above.

All of the automated tasks associated with a Granitic web service endpoint are handled by the [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)
struct.

If you open the file <code>resource/components/components.json</code> you will see:

```json
"artistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "Logic": "ref:artistLogic",
  "PathPattern": "^/artist"
}
```

The component <code>artistHandler</code> is an instance of [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) and we
can define how path binding will work through configuration.

Change the definition of your <code>artistHandler</code> component so it looks like:


```json
"artistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "Logic": "ref:artistLogic",
  "PathPattern": "^/artist/([\\d]+)[/]?$",
  "BindPathParams": ["Id"]
}
```

We've altered the regular expression that this endpoint expects to 

<code>^/artist/([\d]+)[/]?$</code>

so that in order to match an incoming request, the request path _must_ include a number and an optional trailing slash. We've
also defined a [regular expression group](https://www.regular-expressions.info/refcapture.html) around the part of the path that
will be considered as required ID.

We've also added a new field <code>BindPathParams</code> and set it to an array of strings. The number of strings in this
array should match the number of groups in the <code>PathPattern</code> regex. Here we are saying that the first group
should be injected into a field called 'Id'.

The last step is to tell Grantic that the <code>Id</code> field we're refering to is the one on our <code>ArtistRequest</code>
struct. This is done in code by making your <code>ArtistLogic</code> struct implement [handler.WsUnmarshallTarget](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsUnmarshallTarget)

The method <code>UnmarshallTarget() interface{}</code> required by this interface allows each endpoint to create a 'target'
object that any data from a request will be decanted into. 

Change your <code>ArtistLogic</code> struct to look like:

```go
type ArtistLogic struct {
  EnvLabel string
  Log      logging.Logger
}

func (al *ArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

  ar := req.RequestBody.(*ArtistRequest)

  a := new(ArtistDetail)
  a.Name = "Some Artist"

  res.Body = a

  l := al.Log
  l.LogTracef("Request for artist with ID %d", ar.Id)

}

func (al *ArtistLogic) UnmarshallTarget() interface{} {
  return new(ArtistRequest)
}
```

Stop, rebuild and restart your application:

<pre>
grnc-bind && go build && ./recordstore -c resource/config,resource/env/production.json
</pre>

Visiting [http://localhost:8080/artist]() will now result in a 404 Not Found error, but visiting [http://localhost:8080/artist/1234]() should
result in a response and a log line similar to:

<pre>
09/Oct/2017:12:44:13 Z TRACE artistLogic Request for artist with ID 1234
</pre>


### Type assertion

To work flexibly with any custom code you might create, the various methods and interfaces in Granitic's [handler]([handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler))
package tend to work with <code>interface{}</code> types. One of the side effects of this is that the code
in your endpoint's <code>Process</code> method will need to perform a type assertion when
accessing the content's of a request's body. In this case the line

```go
ar := req.RequestBody.(*ArtistRequest)
```

performs the required check.

## Binding query parameters

The technique for binding query parameters to your 'target' object is very similar to that used for binding path parameters. 
Edit your <code>components.json</code> file and add the following to the definition
of your <code>artistHandler</code> component:

```json
   "FieldQueryParam": {
      "NormaliseName": "normalise"
   }
``` 

and add the following Go to the end your <code>ArtistLogic.Process</code> method:

```go
  if ar.NormaliseName != nil && ar.NormaliseName.Bool() {
    a.Name = strings.ToUpper(a.Name)
  }
```

Rebuild and restart your application. Visiting [http://localhost:8080/artist/1234?normalise=true]() will now
cause the returned artist's name to be capitalised.
 

  

