# Application logic

[Reference](README.md) | [Web Services](ws-index.md)

---

A 'logic' component represents the interface between the Granitic framework and your application logic.
Granitic passes your code a [populated](ws-capture.md) and [validated](ws-validate.md) 
[ws.Request](https://godoc.org/github.com/graniticio/granitic/v2/ws#Request) as well as an empty
[ws.Response](https://godoc.org/github.com/graniticio/granitic/v2/ws#Response) into which you can
record any errors your code encounters as well the 'result' data to be encoded in the eventual HTTP response.

## Linking a handler to logic

Every instance of [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsHandler)
requires that you set the `Logic` field to a reference to a component that either:

  * Has a method of the form `ProcessPayload(context.Context, *ws.Request, *ws.Response, *YourStruct) ` 
  * Implements a combination of the [handler.WsRequestProcessor](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsRequestProcessor) and 
  [handler.WsUnmarshallTarget](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsUnmarshallTarget) interfaces.
 
Which of those you use depends on the use cases (listed below) that most closely matches your requirement


### Capture data into an empty struct

The most common use case for web services is that a request will have some information that needs to be 
[captured](ws-capture.md). Generally that information can be parsed into a target struct that does not require any complex
initialisation.

In this case, your logic component should implement a method of the form:

```go
  ProcessPayload(context.Context, *ws.Request, *ws.Response, *YourStruct)
```
                                    
Where `YourStruct` is the type that you want Granitic to instantiate, populate with data from the request and 
pass into your logic component.

### No data to capture

Some web service requests do not require or allow information to be supplied in the HTTP request body, path or
query parameters. 'Health check' or service status endpoints, for example. For these requests your logic
component should implement the [handler.WsRequestProcessor](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsRequestProcessor)
interface, e.g. have a method with the signature:

```go
  Process(ctx context.Context, request *ws.Request, response *ws.Response)
``` 

### Capture data into a complex struct

There are some circumstances under which the [target struct](ws-capture.md) into which Granitic should parse
web service data needs initialisation before it can receive data.

If this is the case for your web service, your logic component must implement 
[handler.WsRequestProcessor](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsRequestProcessor) _and_
[handler.WsUnmarshallTarget](https://godoc.org/github.com/graniticio/granitic/v2/ws/handler#WsUnmarshallTarget).

Before parsing, Granitic will call your logic component's 

```go
  UnmarshallTarget() interface{}
```

method. This method should create a struct, perform whatever initialisation is needed and return a pointer to
that struct. For example:

```go
func UnmarshallTarget() interface{} {
	
 s:= new(MyStruct)

 //Perform intialisation
 
 return s
	
}
```

Your logic component can then recover this later, e.g: 

```go
func Process(ctx context.Context, request *ws.Request, response *ws.Response) {
	
  s := request.RequestBody.(*MyStruct)
	
}
```

## Building a response

Granitic will construct a response based on the state of the [ws.Response](https://godoc.org/github.com/graniticio/granitic/v2/ws#Response)
object when your logic component's `Process...` method returns.

### Body

The contents of the `Response.Body` field will be serialised into the HTTP response body by 
the [ws.ResponseWriter](https://godoc.org/github.com/graniticio/granitic/v2/ws#ResponseWriter) that is set on
the `ResponseWriter` field of your handler. If you are using the [JSONWs](fac-json-ws.md) or [XMLWs](fac-xml-ws.md)
facility this will happen automatically.

You can set this field to any Go value that can be serialised by the [ws.ResponseWriter](https://godoc.org/github.com/graniticio/granitic/v2/ws#ResponseWriter).

### HTTP status code

The HTTP status code (200, 404 etc) that is set on the eventual HTTP response returned to your web service
client is calculated in one of two ways.

  1. Granitic examines the _types_ of errors present in the `Errors` field of the [ws.Response](https://godoc.org/github.com/graniticio/granitic/v2/ws#Response)
      according to the [rules defined here](ws-error.md)
  2. You can explicitly set the desired response code by setting the `HTTPStatus` on the [ws.Response](https://godoc.org/github.com/graniticio/granitic/v2/ws#Response).

---
**Next**: [Error handling](ws-error.md)

**Prev**: [Validation](ws-validate.md)