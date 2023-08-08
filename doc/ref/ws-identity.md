# Request identification

[Reference](README.md) | [Web Services](ws-index.md)

---

It is common practise to generate an unique ID for every HTTP request received by a web service. This ID often takes the
form of a [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) and can be referred to as a request ID or
a correlation ID, among other terms.

These IDs are often generated at the edge of a web architecture in the first server receiving a request from a user 
(a web page or a public API endpoint) then propagated down through calls to internal services. This allows a single
logical request to be traced down through your whole architecture.


Granitic provides an interface [httpserver.IdentifiedRequestContextBuilder](https://godoc.org/github.com/graniticio/granitic/v2/facility/httpserver#IdentifiedRequestContextBuilder)
which allows you to define a component that:

 * Generates new IDs or recovers them from an inbound HTTP request
 * Store that ID in a new [context.Context](https://golang.org/pkg/context/) 
 * Provide a way of extracting an ID from an existing context
 
If you create a component that implements [httpserver.IdentifiedRequestContextBuilder](https://godoc.org/github.com/graniticio/granitic/v2/facility/httpserver#IdentifiedRequestContextBuilder),
it will automatically be injected into your [HTTPServer](fac-http-server.md) using a [decorator](ioc-decorators.md).

## Default request ID

Granitic can automatically generate a request ID (a V4 UUID) for each request. See the [HTTPServer facility documentation](fac-http-server.md)
for more details.

## Accessing the request ID

If your code has access to the [ws.Request](https://godoc.org/github.com/graniticio/granitic/v2/ws#Request) object, it can
use the `ID()` function on that object to recover the string representation of the ID, given a context.

The ID will also be automatically be made available to any [request instrumentation](ws-instrumentation.md) you have set up and,
by using the context key you have used to store the ID in the context, can be logged in [application](fac-logger.md) and
[access](fac-http-server.md) logging.

Other code can access the ID, as long as it have access to the context, by invoking the function `ws.RequestID(context.Context)`


---
**Next**: [Rule based validation](vld-index.md)

**Prev**: [Instrumentation](ws-instrumentation.md)