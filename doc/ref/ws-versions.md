# Version routing

[Reference](README.md) | [Web Services](ws-index.md)

---

It is common practise to allow web service clients to specify the version of an endpoint they want to use on a service, especially
when compatibility breaking changes are made as part of new release of that service. As the representation of version numbers 
and versioning strategies differ substantially between service implementations, Granitic does not provide any concrete types for 
representing versions, extracting requested versions from requests or assessing which version best matches a given handler.

Instead it provides a series of interfaces which, when implemented by your components, allow different instances of
[handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) to be selected to serve 
a request according to the client requested version.

## Extracting a version from the request

You will need to create a component that implements 
[httpendpoint.RequestedVersionExtractor](https://godoc.org/github.com/graniticio/granitic/httpendpoint#RequestedVersionExtractor) 
by defining a method:

```go
func Extract(*http.Request) RequiredVersion
```

Which, given a [http.Request]()


---
**Next**: [Instrumentation](ws-instrumentation.md)

**Prev**: [Identity Access Management](ws-iam.md)