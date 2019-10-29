# Error handling

[Reference](README.md) | [Web Services](ws-index.md)

---

Granitic provides components, patterns and interfaces that allow you to provide a consistent way of communicating errors
to your web service clients. The core types involved in Granitic error handling are the 
[ws.CategorisedError](https://godoc.org/github.com/graniticio/granitic/ws#CategorisedError) and 
[ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors).

These two types are used by both Granitic itself and your application logic.

## Categorised errors

A [ws.CategorisedError](https://godoc.org/github.com/graniticio/granitic/ws#CategorisedError) has four elements:

  * A category to which the error belongs
  * A unique text code to identify the error
  * A human readable message
  * A field name to which the error relates (optional)
  
### Categories

Granitic has a system of categorisation for errors encountered during web service requests. The categories are:

  * Client - the web services client has sent a bad or invalid request
  * Logic - the request was valid, but executing it would violate some form of logical constraint
  * Security - the caller is not allowed to make the request
  * Unexpected - an internal problem (e.g. connectivity outage) was encountered
  
These categories are programmatically represented by constants in the [ws package](https://godoc.org/github.com/graniticio/granitic/ws)

The purpose of the categories is to:

 1. Allow callers to understand where the 'fault' lies for the error
 1. To allow Granitic to choose an appropriate HTTP response code for the web service request
 
### Codes

Associating a code with an error allows callers to recognise and handle types of errors without having to parse a human readable
message. For example, you may decide to use the code `DUPE_EMAIL` for the an error raised because an email address is already
being used by another user.

If you choose to use the [ServiceErrorManager facility](fac-service-errors.md) to manage your error messages in configuration
files, the code is used to find the message associated with the error.

### Message

The message component of the error allows you to provide a more detailed explanations of why the error occurred in a 
human readable form. You can specifiy these messages in code where an error is raised, or use the  
[ServiceErrorManager facility](fac-service-errors.md) to manage your error messages in configuration.

### Fields

When an error specifically relates to a data field provided by the caller, you can (optionally) specify the name of that
field.

## Service Errors

Your web service's [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) creates
a data structure of type [ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors) to collect
errors. It is made available to your [application logic](ws-logic.md) via the `Errors` field of the 
[ws.Response](https://godoc.org/github.com/graniticio/granitic/ws#Response) object passed to your code.

It provides an interface for recording errors as they are encountered. For example:

```go
func (l *Logic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, t *SomeTarget) {

  se := res.Errors
	
  se.AddNewError(ws.Client, "NO_NAME", "You must provide a name")
}
```

### Predefined errors

Using the `AddPredefinedError` method on [ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors)
requires you to have enabled the  [ServiceErrorManager facility](fac-service-errors.md)

## Representing errors in HTTP responses

If no errors are encountered while processing a request, Granitic will set the HTTP status to `200 Okay` and 
use the `Body` field of the [ws.Response](https://godoc.org/github.com/graniticio/granitic/ws#Response) as the HTTP
response body.

If errors have been found, Granitic will examine the categories of the errors to determine the correct response code to
use and replace the body with a data structure specific for communicating errors. For example, if you are using the 
default [JSONWs](fac-json-ws.md) facility with [automatic validation](vld-index.md) enabled, a web service caller might
see something like:

```json
{
  "ByField":{
    "Name":[
     {
       "Code":"C-INVALID_ARTIST",
       "Message":"Cannot create an artist with the information provided."
     }
    ]
  }
}
``` 

The format of error responses is specific to the web services facility, [JSON](fac-json-ws.md) or [XML](fac-xml-ws.md),
you have enabled and is documented there along with instructions on how to change the behaviour to match your project
standards.

### HTTP status codes

Granitic examines the [ws.Response](https://godoc.org/github.com/graniticio/granitic/ws#Response) at the end 
of request processing and chooses the most appropriate HTTP status code according to the  following rules:

  1. If the `Response.HTTPStatus` field is non-zero, use that
  1. If the `Response.Errors.HTTPStatus` field is non-zero, use that
  
Otherwise, if the `Response.Errors` structure:
    
  1. Contains one or more _Unexpected_ errors, use `HTTP 500 - Internal server error`
  1. Contains an _HTTP_ error, convert that error's code to a number and use that
  1. Contains one or more _Security_ errors, use `HTTP 401 - Unauthorized`
  1. Contains one or more _Client_ errors, use `HTTP 400 - Bad Request`
  1. Contains one or more _Logic_ errors, use `HTTP 409 - Conflict`

## Framework errors

Before Granitic passes control to your [logic component](ws-logic.md) any errors encountered while performing
the early phases of request processing are recorded using the [ws.FrameworkError](https://godoc.org/github.com/graniticio/granitic/ws#FrameworkError)
type. 

Generally this process is transparent to your application - errors encountered before your logic are generally due
to malformed client requests that cannot be easily recovered from and Granitic will automatically return a 
`HTTP 400 - Bad Request` response.

But if, you choose, you can disable automatic handling of these errors by setting the `DeferFrameworkErrors`
field of your [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) to `false`.

Your application logic will then have access to the framework errors found via the `FrameworkErrors` field
of the [ws.Request](https://godoc.org/github.com/graniticio/granitic/ws#Request) passed to it.

---
**Next**: [Identity Access Management](ws-iam.md)

**Prev**: [Application logic](ws-logic.md)