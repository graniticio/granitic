# Endpoints and handlers

[Reference](README.md) | [Web Services](ws-index.md)

---
## Endpoints

When a web service request is received by your application, the first step for Granitic is to see if there is an _endpoint_
defined for the request. An endpoint is defined as the combination of the path element of the HTTP request and the HTTP method

For example:

```
  /artist GET
  /artist POST
  /status HEAD 
```

are three different endpoints.

Components that provide endpoints for requests implement the [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider)
interface. 

### Regular expressions

The path that is associated with the endpoint is not expressed as a static string, instead it is specified
as a [Go regular expression](https://golang.org/pkg/regexp/).

Using regular expressions serves two purposes: firstly it allows the endpoint to be less brittle - the regular expression
can be crafted to support mixed-cases or trailing slashes in paths. Secondly, and more importantly, it allows
capture groups to be defined to allow meaningful information to be [extracted from the request path](ws-capture.md). This
is vital for REST-like APIs where IDs are often included as part of paths.

## Handlers

Once Granitic has found an component that defines an endpoint matching the request, it calls the `ServeHTTP` method
on that component. Depending on the implementation of [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider)
in use, that method may be fully implemented by the component in question, or it might defer it to another component.

Whatever the arrangement, Granitic refers to the component which performs the mechanical parts of web service processing 
(parsing, validating, error handling etc) as a _handler_.

For some non-standard endpoints (serving binary data, for example), you may write your own handler, but the vast majority
of handler components you create will use Granitic's built-in [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)

## WsHandler

[WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) supports the bulk of Granitic's
[web processing phases](ws-pipeline.md). It also implements [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider)

A component using [WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) as a type requires:

  * A regex to match a path
  * An HTTP method
  * A reference to (or an inline definition of) a [logic](ws-logic.md) component that implements the interesting work that your web service performs

For example, a minimal [component definition](ioc-definition-files.md) might look like:

```json
"artistHandler": {
  "type": "handler.WsHandler",
  "PathPattern": "^/artist",
  "HTTPMethod": "GET",
  "Logic": {
    "type": "artist.GetLogic"
  }
}

```

[WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) has a number of fields which are
used to customise its behaviour. These customisation options will be explained through the rest of this section.

## Framework errors

Errors encountered by Granitic while serving a web request, but outside of your application code, are referred to as
`framework errors`. The messages that are served to users in those circumstances are defined in configuration. The
default values are:

```json
{
  "FrameworkServiceErrors":{
    "Messages": {
      "UnableToParseRequest": ["PARSE","Unable to parse the body of the request. Please check the content you are sending."],
      "QueryTargetNotArray":  ["QUERYBIND", "Multiple values for query parameter %s. Only one value supported"],
      "QueryWrongType": ["QUERYBIND", "Unable to convert the value of query parameter %s to type %s. Value provided was %s"],
      "QueryNoTargetField": ["QUERYBIND", "No field named %s exists to bind query parameter %s into."],
      "PathWrongType": ["PATHBIND", "Unable to convert the value of a path parameter (group %s) to type %s. Please check the format of your request path. Value provided was \"%s\""]
    },
    "HTTPMessages": {
      "401": "Access to this resource requires authorization.",
      "403": "You do not have permission to interact with that resource.",
      "404": "No such resource.",
      "500": "An unexpected error occurred.",
      "503": "The service is too busy to process your request or is temporarily unavailable."
    }
  }
}
``` 

You can override these messages as you would any other configuration value.

---
**Next**: [Capturing data](ws-capture.md)

**Prev**: [Processing phases](ws-pipeline.md)