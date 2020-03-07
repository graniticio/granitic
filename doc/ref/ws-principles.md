# Web services principles

[Reference](README.md) | [Web Services](ws-index.md)

---

## Complete and configurable processing pipeline

By enabling the [HTTPServer](fac-http-server.md) and one of the [JSONWs](fac-json-ws.md) or [XMLWs](fac-xml-ws.md)
facilities, your service will be ready to serve web requests through a well-defined pipeline of
[processing phases](ws-pipeline.md).

You are free to substitute your own components for those created and managed by Granitic.

## Configurable handler

The core orchestration component in Grantic web service request handling is the [handler](ws-handlers.md). 
Granitic provides a highly configurable default handler [ws.Handler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)
that meets the needs of most (non-binary) web service requests.

## Minimal code boilerplate

Granitic's goal is to minimise the amount of boiler-plate code required to create a web service. Most web services
can be created by declaring a [handler](ws-handlers.md) in your [component definition file](ioc-definition-files.md)
and creating a [logic component](ws-logic.md). 


## Clean separation of framework and application logic

Granitic has a concept of an [application logic](ws-logic.md) component, which represents the boundary from the
HTTP-level processing of a request to your application code. Application code should generally avoid interacting
directly with HTTP-level structures, but Granitic makes them available should your code need them.

## Integration points

Rather than attempt to provide implementations for web service request [identification](ws-identity.md),
[instrumentation](ws-instrumentation.md) and [access management](ws-iam.md), Granitic instead provides well 
defined interfaces and integration points to plug-in your own or third-party solutions.


---
**Next**: [Enabling web services support](ws-enable.md)

**Prev**: [Web services](ws-index.md)