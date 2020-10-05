# Instrumentation

[Reference](README.md) | [Web Services](ws-index.md)

---

It is common to want to integrate a third party instrumentation library or service with web services to monitor performance
or enable cross-service request tracing. Granitic provides interfaces and framework hooks to support this.

There are three steps to supporting instrumentation:

  1. Creating a component that implements [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager)
  1. Having that component return your own implementation of [instrument.Instrumentor](https://godoc.org/github.com/graniticio/granitic/instrument#Instrumentor) 
  1. Configuring the [HTTPServer facility](fac-http-server.md) to enable instrumentation
  
The first two steps are explained below, but [configuration of the HTTPServer facility is documented here]((fac-http-server.md)).  

## Request Instrumentation Manager  

The role of the [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager)
is to allow Granitic to start instrumentation of a request, obtain an interface for recording discrete events within the 
request and access a method for stopping instrumentation.

You component must implement the method:

```go
func Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, Instrumentor, func())
```

### Contexts

Although optional, it is recommended that your implementation of `Begin` stores the `Instrumentor` it creates in a new context
using the Granitic function `instrument.AddInstrumentorToContext`. This means that any code in your application can retrieve
the `Instrumentor` using the corresponding `instrument.InstrumentorFromContext` method, or cleanly start new instrumentation
events using the `instrument.Event` and `instrument.Method` methods.

### End function

Your begin method is required to return a function (most likely a closure) that can be called by Granitic to end instrumentation
for a request.

## Instrumentor

Your implementation of [instrument.Instrumentor](https://godoc.org/github.com/graniticio/granitic/instrument#Instrumentor)
allows sub-sections of a request (events) to be instrumentated. For example, you may want to separately record the execution
time of each method used by during your request.

Your code can interact directly with the `Instrumentor` or indirectly using the helper functions `instrument.Event` 
and `instrument.Method` methods.

## Handling goroutines

If your underlying instrumentation framework supports it, Instrumentor provides hooks to allow a child Instrumentor to
be spawned for a new goroutine (`Fork`) and the results merged back into the overall request instrumentation (`Integrate`).

If you do not support this feature, `Integrate` should just return the same `Instrumentor` it is passed and `Integrate` should
do nothing.

## Gaining access to additional data

Instrumentation starts as soon as the [HTTPServer facility](fac-http-server.md) receives an HTTP request - before anything is known
about the request. As your `Instrumentor` may need access to the additional information (the handler that processes the request, the ID associated
with the request), Granitic automatically calls the 

```go
Amend(additional Additional, value interface{})
``` 

method as new data is available. Your code must explicitly convert the `interface{}` value passed into `Amend` according to the value of 
the [instrument.Additional](https://godoc.org/github.com/graniticio/granitic/instrument#Additional) pseudo-enum.

### Manually supplying additional data

If your application needs to supply additional data about an event after it is has been created, you can use
the `instrument.Amend` helper function. This will find your `Instrumentor` and call its `Amend` function with
`Custom` as the `additional` parameter.

## Ending instrumentation

Granitic will call the end function created by the `Begin` method on your [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager)
when the HTTP request that is being instrumented is closed (more specifically when the `httpserver.HTTPServer`'s `handleAll` method exits).

It is highly recommended that any _flushing_ actvitiy (sending the instrumentation data to storage) is handled asynchronously
when the end function is called, otherwise the underlying HTTP request from your client's web service call will not close
until the flushing is complete.

It is also highly recommended that your  [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager)
implements as many of Granitic's [lifecycle interfaces](ioc-lifecycle.md) as is appropriate, so that connections to 
your instrumentation service are created during application startup and pending data is not lost when your application
shuts down.

  
---
**Next**: [Request identification](ws-identity.md)

**Prev**: [IAM](ws-iam.md)