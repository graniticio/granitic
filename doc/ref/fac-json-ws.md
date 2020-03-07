# JSON Web Services (JSONWs)
[Reference](README.md) | [Facilities](fac-index.md)

---

Enabling the JSON web services facility automatically injects components into your [web service handlers](ws-handlers.md)
such that HTTP request bodies are parsed as JSON documents and HTTP response bodies are formatted as JSON documents.

## Enabling

The JSONWs facility is _disabled_ by default. To enable it, you must set the following in your configuration

```json
{
  "Facilities": {
    "JSONWs": true
  }
}
```

### Prerequisites

In order to use the JSONWs service, you must also enable the [HTTPServer facility](fac-http-server.md)

## Configuration

The default configuration for this facility can be found in the Granitic source under `facility/config/jsonws.json`
and is:

```json
{
  "JSONWs":{
    "ResponseWriter": {
      "DefaultHeaders": {
        "Content-Type": "application/json; charset=utf-8"
      },
      "IncludeRequestID": false,
      "RequestIDHeader": "request-id"
    },
    "Marshal": {
      "PrettyPrint": false,
      "IndentString": "  ",
      "PrefixString": ""
    },
    "WrapMode": "BODY",
    "ResponseWrapper": {
      "ErrorsFieldName": "Errors",
      "BodyFieldName":   "Response"
    }
  }
}
```

### Headers

`JSONWs.ResponseWriter.DefaultHeaders` contains HTTP response headers that will automatically added to every HTTP
response served by your application. The `Content-Type` header is by default set to  `application/json; charset=utf-8`
which is the most common content type and encoding for JSON.

You may add additional headers to the response by setting them in your configuration, e.g:

```json
{
  "JSONWs":{
    "ResponseWriter": {
      "DefaultHeaders": {
        "My-Common-Header": "my value"
      }
    }
  }
}
```

#### Request ID Header

If you have enabled [request identification](ws-identity.md), you can have the ID associated with a request included as a response
header by setting `JSONWs.ResponseWriter.IncludeRequestID` to `true`. The name of the header can be controlled by
changing the value of `JSONWs.ResponseWriter.RequestIDHeader`.

### JSON formatting

The value of your [ws.Response.Body](https://godoc.org/github.com/graniticio/granitic/ws#Response) field will be marshalled
into JSON using Go's standard [json.Marshal](https://golang.org/pkg/encoding/json/#Marshal) function.

If you set `JSONWs.Marshal.PrettyPrint` to `true`, [json.MarshalIndent](https://golang.org/pkg/encoding/json/#MarshalIndent)
will be used instead and the configuration values for `IndentString` and `PrefixString` will be passed into
that function.

### Response wrapping

By default, Granitic will use the JSON representation of your [ws.Response.Body](https://godoc.org/github.com/graniticio/granitic/ws#Response)
field as the entire HTTP response body. If an [error](ws-error.md) is found, the entire HTTP body is replaced with
the error document structure described below.

Some applications prefer to have a consistent structure regardless of whether the request was successful or not.
That behaviour can be enabled in Granitic by setting `JSONWs.WrapMode` to `WRAP`. This means all requests will be
wrapped with:

```json
{
  "Response": { },
  "Errors": {}
}
```

`Response` will only be populated if the request was successful and `Errors` will only be populated if a problem was
found. The labels `Response` and `Errors` can be modified by changing the `JSONWs.ResponseWrapper.ErrorsFieldName` and
`JSONWs.ResponseWrapper.BodyFieldName` configuration.

## Behaviour

Enabling this facility causes several components to be created and automatically injected into any [handlers](ws-handlers.md)
that you have defined that are of the type [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler).

### ResponseWriter

Your handler's `ResponseWriter` field will be set to an instance of [ws.MarshallingResponseWriter](https://godoc.org/github.com/graniticio/granitic/ws#MarshallingResponseWriter).
This type is agnostic of the serialisation format, so it is customised for JSON serialisation using the
types defined in the [ws/json package](https://godoc.org/github.com/graniticio/granitic/ws/json).

Together these types are responsible for populating the body, headers and status code of the HTTP responses
to your webservice request.

The ResponseWriter is also injected into the [HTTP server](fac-http-server.md) so that it can handle requests
than are not matched to a handler (not found, too busy etc).

### Unmarshaller

Your handler's `Unmarshaller` field will be set to an instance of [json.Unmarshaller](https://godoc.org/github.com/graniticio/granitic/ws/json#Unmarshaller),
which is a simple wrapper over Go's built-in JSON decoding functions.

## Customisation

Granitic will not inject the above components into your handlers if the relevant target field is already populated. 
So if you explicitly set an `Unmarshaller` or `ResponseWriter` with a reference to one of your own components in 
your [component definition](ioc-definition-files.md)  (or in a [component template](ioc-templates.md)) the `JSONWs`
facility will not overwrite it. 

## Changing default HTTP status codes

The set of HTTP status codes used when an error is found, [according to the rules here](ws-error.md), are defined in 
configuration. The default values are:

```json
{
  "WS": {
    "HTTPStatus": {
      "NoError": 200,
      "Client": 400,
      "Security": 401,
      "Unexpected": 500,
      "Logic": 409
    }
  }
}
```

You can change one or more of these codes by overriding the value in your application's configuration.

### Advanced customisation

If you want to tweak the behaviour of the components made available by this facility (especially the `ResponseWriter`)
you can make use of the advanced [framework modification](ioc-definition-files.md) feature to inject your own components
into fields of the components listed below.

## Component reference

The following components are created when this facility is enabled:

| Name | Type |
| ---- | ---- |
| grncJSONResponseWriter | [ws.MarshallingResponseWriter](https://godoc.org/github.com/graniticio/granitic/ws#MarshallingResponseWriter) |
| grncJSONUnmarshaller | [json.Unmarshaller](https://godoc.org/github.com/graniticio/granitic/ws/json#Unmarshaller) |

---
**Next**: [XML Web Services](fac-xml-ws.md)

**Prev**: [Logger](fac-logger.md)