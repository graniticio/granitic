# Enabling web services support

[Reference](README.md) | [Web Services](ws-index.md)

---

In order to handle HTTP web service requests, your application must enable two facilities: the [HTTPServer facility](fac-http-server.md)
and _either_ the [JSONWs facility](fac-json-ws.md) or the [XMLWs facility](fac-xml-ws.md) in one of its [configuration files](cfg-files.md).

```json
{
  "Facilities": {
    "HTTPServer": true,
    "JSONWs": true
  }
}
```

or

```json
{
  "Facilities": {
    "HTTPServer": true,
    "XMLWs": true
  }
}
```

Granitic will then automatically detect any [components](ioc-definition-files.md) that implement 
[httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider) and route HTTP requests
to them according to their URI and HTTP method.

In practise, you will generally define components of type [handler.WsHandler]((https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler))
which implements [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider). Custom
implementations of [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider) are covered
in [handlers and endpoints](ws-handlers.md)

Granitic's web service implementation is designed to minimise the amount of boilerplate code you need to write to handle
common parts of the web service request processing lifecycle (URI matching, parsing, validation etc). Instead it provides 
a well defined set of [processing phases](ws-pipeline.md) and components which you can customise or replace.   
 
---
**Next**: [Processing phases](ws-pipeline.md)

**Prev**: [Web service principles](ws-principles.md)