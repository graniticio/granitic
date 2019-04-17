# Web service processing phases

Back to: [Reference](README.md) | [Web Services](ws-index.md)

---

In Granitic, the components and processes that are used to service an HTTP request to your application are organised 
into a series of 'phases'.  Although the order of these phases is fixed, most of them can be customised, disabled 
or even swapped out for your own implementation.

A request is initially handled by Granitic's [HTTP server](fac-http-server.md) facility then passed on to a component
that implements [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider) and
supports the incoming request's URL and HTTP method. In the vast majority of cases that component will be an instance of
Granitic's built-in [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler).

Your application will associate a [logic](ws-logic.md) component with each handler to perform business logice and build
a response.

## Phases


| Component        | Step           | Purpose  |
| ------------- |-------------| -----|
| [HTTP Server](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#HTTPServer)| Instrumentation | Begin timing/metric gathering if your application supports [instrumentation](ws-instrumentation.md) |
| | Load management      |   Reject request if too busy/suspended |
| | Request identification      | Assign a unique ID to the request if [supported by your application](ws-identity.md) |
| | Matching| Attempt to match the request's path and HTTP method to a [handler](ws-handlers.md) |
| | Version routing | If your application supports [version routing](ws-versions.md) the correct handler for the request version is found|
| [Handler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) | Authentication | If your application supports [identity and access management](ws-iam.md) the request caller is authenticated
| | Authorisation| If the caller has been authenticated, their permission to access this functionality is checked|
| | Data extraction | Data from the request body, path and query parameters is [parsed into a struct](ws-capture.md)|
| | Validation | The extracted data is [validated](ws-validate.md) to confirm that it is ready to be processed|
| [Application logic](ws-logic.md) | Processing | The [application logic](ws-logic.md) component performs the actual function of the web service |
| | Response construction | A Go struct is populated with data that will form the response to the request |
| [Handler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) | Response serialisation| Serialise the response struct (normally to [JSON](fac-json-ws.md) or [XML](fac-xml-ws.md)) |
| [HTTP Server](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#HTTPServer) | Request finalisation | Instrumentation ended and HTTP request closed|

### Error handling

The above represents the 'happy path' for a request. Most of the phases can terminate early triggering automatic [error handling](ws-error.md)

---
**Next**: [Endpoints and handlers](ws-handlers.md)

**Prev**: [Enabling web services](ws-enable.md)