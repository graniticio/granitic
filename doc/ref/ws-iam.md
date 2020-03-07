# Identity and Access Management (IAM)

[Reference](README.md) | [Web Services](ws-index.md)

---

Identity and Access Management (IAM) refers to the process of confirming the identity of the user of a system and then
using that identity to determine which functionality within a system that user can access. For web services, this generally
means that some [endpoints](ws-handlers.md) are only available to certain users or that the behaviour of those endpoints
changes according to the user.

Granitic does not offer any functionality for managing users, roles or permissions but provides a number of hooks into
which you can plug-in your own IAM implementations (or interact with a third party IAM solution).

## Handler fields

[handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) provides a number of fields
you can use to configure the IAM behaviour of your endpoints. These are all covered in detail below:

```go

// A component able to examine a request and see if the caller is allowed to access this endpoint.
AccessChecker ws.AccessChecker

// Check caller's permissions after request has been parsed (true) or before parsing (false).
CheckAccessAfterParse bool

// Whether on not the caller needs to be authenticated (using a ws.Identifier) in order to access the logic behind this handler.
RequireAuthentication bool

// A component that can examine a request to determine the calling user/service's identity.
UserIdentifier ws.Identifier 
```


## Identifying and authenticating a caller

A component implementing [ws.Identifier](https://godoc.org/github.com/graniticio/granitic/ws#Identifier) is responsible
for examining an HTTP request to determine who is making the web service call. The information required to do this is
generally (but not always) encoded in HTTP request headers.

You component needs to have a method:

```go
 func Identify(ctx context.Context, req *http.Request) (iam.ClientIdentity, context.Context)
```

And you make this component available to your [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)
by setting a reference to it via the `UserIdentifier` field.

### ClientIdentity

Your `Identify` method returns an instance of [iam.ClientIdentity](https://godoc.org/github.com/graniticio/granitic/iam#ClientIdentity)
which should be set up to reflect the current state of the user - whether or not the user has been authenticated and 
providing some string representation of the user's ID that can be used in application and access log files.

As [iam.ClientIdentity](https://godoc.org/github.com/graniticio/granitic/iam#ClientIdentity) is of type `map[string]interface{}`
you can also store any data you like about the user, which your application code can retrieve later. The `ClientIdentity` 
is passed into your [logic component](ws-logic.md) as part of the [ws.Request](https://godoc.org/github.com/graniticio/granitic/ws#Request)

### Context

If your application makes further calls to downstream services, it is likely that the information identifying the user 
will need to be propagated to those services. The Go pattern for transparently moving meta-data about a request through
your application is via a [context.Context](https://golang.org/pkg/context/).

If you require this functionality, it is recommended that your `Identify` returns a new context containing enough
information to recreate the HTTP encoded representation of a user identity when the call to a downstream service is made.

## Requiring authentication

You can require a user to be authenticated to use an endpoint. If you set the `RequireAuthentication` field to `true`
on your [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler), Granitic will automatically
return a `401 Unauthorized` HTTP response code if the [iam.ClientIdentity](https://godoc.org/github.com/graniticio/granitic/iam#ClientIdentity) 
you constructed in your `Identify` method has had it's `SetAuthenticated` method set to `false`.


## Checking authorisation

Once a user has been identified, you can optionally provide a component to check if they are authorised to access the
current endpoint. You should create a component that implements [ws.AccessChecker](https://godoc.org/github.com/graniticio/granitic/ws#AccessChecker)
and set a reference to it on the `AccessChecker` field of your [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)

Your component needs to implement the method:

```go
func Allowed(ctx context.Context, r *Request) bool
```

And return `false` if the user is not allowed to access the current endpoint, which will result in a `403 Forbidden` HTTP
response code being sent to the caller.

### Authorise after parse

By default, the authorisation check occurs before the body of the inbound request is [parsed](ws-capture.md). If your
code needs access to the data embedded in the HTTP request's body to determine whether a user is authorised or not, you 
can set the `CheckAccessAfterParse` field on your handler to `true`. If the body is successfully parsed, it will be
available in the [ws.Request](https://godoc.org/github.com/graniticio/granitic/ws#Request) passed into your `Allowed`
method.

Be aware that this is potentially a security risk - callers can pass in invalid bodies in an attempt to use the resulting
errors to determine the structure of a valid request.

---
**Next**: [Instrumentation](ws-instrumentation.md)

**Prev**: [Error handling](ws-error.md)