# Service Error Manager (ServiceErrorManager)
[Reference](README.md) | [Facilities](fac-index.md)

---

Enabling the Service Error Manager facility allows human readable error messages and the categorisation of those errors 
to be managed in configuration files. This simplifies application logic where errors can
be raised using the [ws.ServiceErrors.AddPredefinedError](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors)
method without having to manage messages and categories in Go source code.

This facility must be enabled to make use of [automatic validation](vld-index.md).

## Enabling

The ServiceErrorManager facility is _disabled_ by default. To enable it, you must set the following in your configuration

```json
{
  "Facilities": {
    "ServiceErrorManager": true
  }
}
```

## Configuration

The default configuration for this facility can be found in the Granitic source under `facility/config/servicerror.json`
and is:

```json
{
  "ServiceErrorManager":{
    "PanicOnMissing": true,
    "ErrorDefinitions": "serviceErrors"
   }
}
```

### Configuration location

By default, definitions for errors (see below) are expected to be located at the configuration path `serviceErrors`. This can be
overridden with an alternative location by setting `ServiceErrorManager.ErrorDefinitions` to your preferred location.

If no definitions are found when your application starts and `ServiceErrorManager.PanicOnMissing` is set to `true`
(the default), your application will shutdown cleanly with an error.


## Defining messages

A Granitic web service error is represented in code as a [CategorisedError](https://godoc.org/github.com/graniticio/granitic/ws#CategorisedError) 
which has three common components a category, a code and a message. You can define these in configuration using the following
format (a two-dimensional string array):

```go
"serviceErrors": [
  ["C", "INVALID_ARTIST", "Cannot create an artist with the information provided."],
  ["C", "NAME_MISSING", "You must supply the Name field on your submission."]
]
```

### Category

The first element of each error is a [category](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrorCategory). 
The category affects the [resulting HTTP status code](https://godoc.org/github.com/graniticio/granitic/ws) used when returning a web service response that contains one or
more errors.

The valid categories for errors defined in configuration are:

| Code | Name | Description |
| ---- | ---- | ----------- |
| U | Unexpected | A server side problem that the web service client could not have foreseen |
| C | Client | A problem with the data submitted by the web service client that it should have foreseen |
| L | Logic | A violation of 'business' logic |
| S | Security | Unauthenticated or unauthorised access |
 
### Code

A code is short text identifier for a particular error that must be unique across all error messages defined in configuration.
The code may be used in application logic to raise an error against a request using the 
[ws.ServiceErrors.AddPredefinedError](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors) method.

### Message

The message is the text associated with the error that will be included in the response body sent back to web 
service clients.

## Missing error detection

Granitic components that make use of the service error manager (e.g. [automatic validation](vld-index.md)) automatically
announce which error codes they are using. If no corresponding message is found in configuration an error will be logged
at application startup.

If you want your own components to be able to announce which codes they are using they should implement
[grncerror.ErrorCodeUser](https://godoc.org/github.com/graniticio/granitic/grncerror#ErrorCodeUser)


## Component reference

The following components are created when this facility is enabled:

| Name | Type |
| ---- | ---- |
| grncServiceErrorManager | [grncerror.ServiceErrorManager](https://godoc.org/github.com/graniticio/granitic/grncerror#ServiceErrorManager) |

---
**Next**: [Runtime Control](rtc-index.md)

**Prev**: [Runtime Control facility](fac-runtime.md)