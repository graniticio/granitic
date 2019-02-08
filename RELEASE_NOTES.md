# Release notes

## 1.3.0  (2018-xx-xx)

## General
  
  * Reference documentation for the framework now available in `doc/ref`
  * YAML support for configuration and component definition files via [graniticio/granitic-yaml](https://github.com/graniticio/granitic-yaml)
  * Improved test coverage
  
## IOC

### Empty Struct factory functions

The IOC container supports a new prefix `empty-struct:` or `es:`. This is used in conjunction with the name of a struct e.g.

```json
    "YourField": "es:yourpackage.YourStruct"
```

This will cause an anonymous function:

```go
    func() interface{} {
        return new(yourpackage.YourStruct)
    }    
```

to be injected into `YourField`, so that field must be of type `func() interface{}`. `yourPackage` must be present in the 
`packages` section at the top of your components file.

See below for an example of where this technique is useful.

## Web services

### Instrumentation

Granitic now supports a pattern for instrumenting web service requests. You can provide implementations
of `instrument.RequestInstrumentationManager` and `instrument.Instrumentor` to integrate with whichever tool
or provider you use for collating instrumentation data. Refer to the `instrument` and `facility/httpserver` package documentation for more details.

### Alternative to implementing WsUnmarshallTarget

Previously the `Logic` component attached to each instance of `ws.WsHandler` was required to implement `ws.WsUnmarshallTarget` to create an
empty struct for unmarhsallling HTTP body, path and parameter data into. This can now be streamlined by setting:

```json
  "CreateTarget": "es:yourpackage.YourStruct"    
```

In the component definition for your handler, where `yourpackage.YourStruct` is the type that your `UnmarshallTarget()` method would've returned.
`yourPackage` must be present in the `packages` section at the top of your components file.
 


### Alternative to implementing WsRequestProcessor

Previously the `Logic` component attached to each instance of `ws.WsHandler` was required to implement `ws.WsRequestProcessor`. 
You can now use any struct as the Logic component as long as it implements `ws.WsRequestProcessor` _OR_ has a method like:

```go
  ProcessPayload(ctx context.Context, request *ws.WsRequest, response *ws.WsResponse, payload *YourStruct)
```

Where `YourStruct` is the same type returned by your Logic component's `UnmarshallTarget()` method, e.g:

```go
  ProcessPayload(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse, ar *ArtistRequest)
```

This approach means that your logic code can immediately use data extracted from the incoming HTTP request instead of having
to use the pattern:

```go
  sar := req.RequestBody.(*SubmittedArtistRequest)
``` 

### Bug fixes

 * Component definition files containing a slice of `float64` could not be bound if the first element was parseable as an int.
 * The `ServiceErrorManager` was not respecting the value of `ErrorCodeUser.ValidateMissing()` when complaining about missing error codes.
 * Using `ConfigAccessor` to try and push configuration into an unsupported type of target field was not returning an error.
 * Some configuration parsing errors were causing Granitic to exit rather than return an error
 
## 1.2.1  (2018-10-08)

 * GoDoc improvements
 * Tests now run if ```GRANITIC_HOME``` not set
 * Compatible with ```fmt.xprintx``` checks in ```go vet```

## 1.2.0  (2018-07-25)

### General
 * Now requires Go 1.9 or later
 
### Logging
 * Default log prefix now wraps component names in square brackets to improve readability
 
### Build and configuration

Applications built with Granitic no longer need access to a copy of Granitic's built-in JSON configuration files at runtime. This means that you do not 
have to set the GRANITIC_HOME environment variable on the servers you are deploying your applications to.

### Scheduled tasks

The new `TaskScheduler` facility allows processes to be run at predetermined intervals. Includes support for

 * Natural language definition of intervals (every `2 days at 1405` etc)
 * Control over error and retry behaviour
 * Notification to other components when tasks end (successfully or not)
 * Control over overlapping task runs
 
Refer to the [granitic.schedule GoDoc](https://godoc.org/github.com/graniticio/granitic/v2/schedule) for more details.

### RDBMS

Improved support for connecting to multiple databases using the `RdbmsAccess` facility. Refer to the [granitic.rdbms GoDoc](https://godoc.org/github.com/graniticio/granitic/v2/rdbms) 
for more information.

### JSON Web Services

__BREAKING CHANGE__

 By default, the JsonWs facility now no longer wraps response objects in a wrapper with 'response' and 'errors' sections.
 Instead the object returned by the WsHandler is serialised as the HTTP response body unless errors are present, in which case
 the errors structure is used as the response body.
 
 To revert to the previous behaviour, set:
 
 ```json
 {
   "JsonWs": {
     "WrapMode": "WRAP"
   }
 }
 ``` 
 
 in your application's configuration

## 1.1.0  (2018-01-11)

### General
 * Various internal changes to make use of updated libraries in Go 1.8
 * Various changes of all-caps abbreviations (e.g. ID, HTTP) to  Mixed Caps (Id, Http) improve readability of method and
 type names

### RDBMS

 * Streamlined [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/v2/rdbms) methods (xxxTagxxx methods removed, xxxParamxxx methods more flexible with the type of arguments they accept)
 * [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms) interface no longer requires DatabaseFromContext - implementations can now implement the optional ContextAwareDatabaseProvider interface
 * [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms) interface no longer requires InsertIDFunc - implementations can now implement the optional NonStandardInsertProvider interface
 * If RdbmsClient was created with a context, all operations on the underlying [sql.Db](https://golang.org/pkg/database/sql/#Db) object use the Context variants of methods where possible
 * Transactions can now be opened with [sql.TxOptions](https://golang.org/pkg/database/sql/#TxOptions)
 * Proxy interfaces introduced in front of [sql.Db](https://golang.org/pkg/database/sql/#Db) and [sql.TxOptions](https://golang.org/pkg/database/sql/#TxOptions) to facilitate testing.
 
## QueryManager
 
 * Escaping of parameter values and handling of missing values now deferred to new [ParamValueProcessor](https://godoc.org/github.com/graniticio/granitic/v2/dsquery#ParamValueProcessor) components.
 * Two built-in [ParamValueProcessor](https://godoc.org/github.com/graniticio/granitic/v2/dsquery#ParamValueProcessor) components are available - select by setting QueryManager.ProcessorName to <code>configurable</code> or <code>sql</code>
 * Default is <code>configurable</code> which mimics Granitic 1.0 behaviour.
 * Choosing <code>sql</code> will set missing parameter values to <code>null</code> and map bools to configurable DB specific values.
 * Parameters in a query can now be marked as required by prefixing their name with <code>!</code> in the query template.

### HttpServer

 * Contexts passed to WsHandler are now inherited from the context on http.Request
 * Server makes use of new Shutdown and Close http.Server methods during framework shutdown


### Validation

 * validate.ExternalXXXValidator interfaces (i.e ExternalInt64Validator) now return an error as well as a bool
 
### Fixes

 * [Issue 001](https://github.com/graniticio/granitic/v2/issues/1) - Nilable fields on target objects now initialised if query binding in use
 * [Issue 002](https://github.com/graniticio/granitic/v2/issues/2) - STOPALL validation operation now works as intended
 
<hr/> 

## 1.0.1 (2017-08-09)

 * Windows support
 * Uses Go context package (rather than golang.org/x/net/context)
 * Now requires Go 1.7

<hr/> 

## 1.0.0 (2016-10-13)

Initial release
 
