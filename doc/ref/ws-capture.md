# Capturing data

[Reference](README.md) | [Web Services](ws-index.md)

---

Aside from the simplest `GET` and `HEAD` requests, most web service requests include data to be processed and/or stored. 
This data can be included in four parts of the HTTP request:

 * The request body
 * The request path (the part of the URL after the domain and before the `?` symbol)
 * Query parameters (the name/value pairs after the `?` symbol)
 * Request headers
 
Granitic provides functionality to automatically capture path, query parameter and body data and parse it into any Go struct
that you nominate, assuming that you are using an instance of [ws.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)
as your handler component.

## Data capture target

The struct that you nominate to receive data is referred to as the _target_ or _target object_. An instance will be automatically
be created by your [handler](ws-handlers.md). There are number of ways in which your code can tell Granitic _which_ type to instantiate by
implementing interfaces or specific method signatures on your [logic component](ws-logic.md). The [logic component documentation](ws-logic.md)
explains this in detail.

## Request body parsing

Data encoded in the HTTP request's body is parsed by a component implementing [ws.Unmarshaller](https://godoc.org/github.com/graniticio/granitic/ws#Unmarshaller)
that is injected into the `Unmarshaller` field on your [handler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler).

If you have enabled the [JSONWs](fac-json-ws.md) or [XMLWs](fac-xml-ws.md) facility, Granitic will automatically inject 
an `Unmarshaller` into your handler.

### Mapping body data to struct fields

How data in the body is mapped to fields in your target struct depends on the implemention of [ws.Unmarshaller](https://godoc.org/github.com/graniticio/granitic/ws#Unmarshaller).
The built-in `Unmarshaller`s for JSON and XML are documented in the  JSONWs](fac-json-ws.md) and [XMLWs](fac-xml-ws.md) 
facility documentation.

### Providing a custom Unmarshaller

A common pattern for web-services is that the majority of endpoints on a service support a single text-based standard for
request and response bodies (e.g. all JSON). It is also common for a small number of endpoints to use a different standard
(e.g. receiving binary or HTML form encoded data). For those cases you can write your own component implementing 
[ws.Unmarshaller](https://godoc.org/github.com/graniticio/granitic/ws#Unmarshaller) and explicit inject them into
your handler in your [component definition files](ioc-definition-files.md).


### Errors during parsing

If the request body cannot be parsed into your target object, a [FrameworkError](https://godoc.org/github.com/graniticio/granitic/ws#FrameworkError)
will be recorded. Framework errors are explained in the web service [error handling documentation](ws-error.md) documentation,
but the practical effect is the the client will receive an `HTTP 400` response.

## Path binding

Extracting information from a request's path and injecting it into your target object is known as _path binding_. Path
binding allows REST-like behaviour where information about resources to read or manipulate are embedded in the request
path.

Consider a request:

`/artist/12/album/2 GET`

There are two IDs embedded in the path. Assuming we have target object like:

```go
type AlbumQuery struct {
	ArtistID int
	AlbumID int
}
```

We can set up our handler to capture path information into the target object like:

```json
"getAlbumHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "PathPattern": "^/artist/([\\d]+)/album/([\\d]+)",
  "BindPathParams": ["ArtistID", "AlbumID"]
}
```

### Capture groups

The regular expression for the endpoint URL on this handler (with JSON escaping removed) is:

`^/artist/([\d]+)/album/([\d]+)`

Which defines two regular expression capture groups in brackets. Both capture groups in this example match a sequence of
digits.

The value of the field `BindPathParams` is a list of field names on the target object. The order is significant - the
first field in the list will be populated with the value from the first capture group and so on.

#### Using regular expressions to enforce type safety

In the example above, the handler will _only_ match the client's request if the values provided in the path are integers.
The client would just receive an `HTTP 404` response if they requested: `/artist/-12/album/true`, for example. It is 
recommended that you adopt this practise.

## Query parameter binding

Query parameters are the name-value pairs after the `?` separator in the request URL.

Consider a request:

`/artist-album?artistID=12&albumID=2`

Assuming the same target object from the path binding example, you have two options when defining a handler

### Auto-binding

```json
"getAlbumHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "PathPattern": "^/artist-album,
  "AutoBindQuery": true
} 
```

With `AutoBindQuery` set to `true`, the values of query parameters will be injected into fields on your target object
where the field names _exactly_ (case sensitive) match the query parameter name.


### Explicit binding

```json
"getAlbumHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "PathPattern": "^/artist-album,
  "FieldQueryParam": {
    "ArtistID": "ArtistID",
    "AlbumID": "AlbumID"
  }
}
```

`FieldQueryParam` is a string/string map where the keys are _field names_ on the target object and the values are the
_names of query parameters_. This approach allows you define names of query parameters that break the rules of Go
struct field names e.g:

```json
"FieldQueryParam": {
  "ArtistID": "artist-id",
  "AlbumID": "album-id"
}
```

#### Missing values

If a query parameter is not provided, no error will be thrown. This scenario should be handled by [validation](ws-validate.md)

#### Incorrect types

If a query parameter is provided where the value is incompatible with the type of the field on the target object, a 
[framework error](ws-error.md) will be raised. 


## Path and query supported types

Path and query binding supports the same set of types to parse data into. These can be:

  * Any Go basic type (except `uintptr`, `complex64` and `complex128`)
  * Any of Granitic's [nilable struct types](ws-nilable.md)
  * A slice of any of the above
  
For slices, the query/path parameter value should be a comma delimited list of values.

For numeric types a [framework error](ws-error.md) will be raised if the parameter value is numeric, but doesn't fit
into the target type.

## HTTP request headers

There is no explicit support in Granitic for binding HTTP request headers to target objects. You may choose to allow your
[logic component](ws-logic.md) to have access to the headers (and the underlying HTTP request and response objects) by
setting `AllowDirectHTTPAccess` to `true` on your handler.

There are integration points for [IAM](ws-iam.md), [instrumentation](ws-instrumentation.md), [versioning](ws-versions.md)
and [identification](ws-identity.md) where you will have access to HTTP request headers without having to set
`AllowDirectHTTPAccess` to `true`. 


---
**Next**: [Nilable types](ws-nilable.md)

**Prev**: [Endpoints and handlers](ws-handlers.md)