# Release notes

## 1.2.0  (2018-xx-xx)

### General
 * Now requires Go 1.9 or later
 
### Logging
 * Default log prefix now wraps component names in square brackets to improve readability
 
### Build and configuration

Applications built with Granitic no longer need access to a copy of Granitic's built-in JSON configuration files at runtime. This means that you do not 
have to set the GRANITIC_HOME environment variable on the servers you are deploying your applications to.

### RDBMS

Improved support for connecting to multiple databases using the RdbmsAccess facility. Refer to the [granitic.rdbms GoDoc](https://godoc.org/github.com/graniticio/granitic/rdbms) 
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

 * Streamlined [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms) methods (xxxTagxxx methods removed, xxxParamxxx methods more flexible with the type of arguments they accept)
 * [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms) interface no longer requires DatabaseFromContext - implementations can now implement the optional ContextAwareDatabaseProvider interface
 * [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms) interface no longer requires InsertIDFunc - implementations can now implement the optional NonStandardInsertProvider interface
 * If RdbmsClient was created with a context, all operations on the underlying [sql.Db](https://golang.org/pkg/database/sql/#Db) object use the Context variants of methods where possible
 * Transactions can now be opened with [sql.TxOptions](https://golang.org/pkg/database/sql/#TxOptions)
 * Proxy interfaces introduced in front of [sql.Db](https://golang.org/pkg/database/sql/#Db) and [sql.TxOptions](https://golang.org/pkg/database/sql/#TxOptions) to facilitate testing.
 
## QueryManager
 
 * Escaping of parameter values and handling of missing values now deferred to new [ParamValueProcessor](https://godoc.org/github.com/graniticio/granitic/dsquery#ParamValueProcessor) components.
 * Two built-in [ParamValueProcessor](https://godoc.org/github.com/graniticio/granitic/dsquery#ParamValueProcessor) components are available - select by setting QueryManager.ProcessorName to <code>configurable</code> or <code>sql</code>
 * Default is <code>configurable</code> which mimics Granitic 1.0 behaviour.
 * Choosing <code>sql</code> will set missing parameter values to <code>null</code> and map bools to configurable DB specific values.
 * Parameters in a query can now be marked as required by prefixing their name with <code>!</code> in the query template.

### HttpServer

 * Contexts passed to WsHandler are now inherited from the context on http.Request
 * Server makes use of new Shutdown and Close http.Server methods during framework shutdown


### Validation

 * validate.ExternalXXXValidator interfaces (i.e ExternalInt64Validator) now return an error as well as a bool
 
### Fixes

 * [Issue 001](https://github.com/graniticio/granitic/issues/1) - Nilable fields on target objects now initialised if query binding in use
 * [Issue 002](https://github.com/graniticio/granitic/issues/2) - STOPALL validation operation now works as intended
 
<hr/> 

## 1.0.1 (2017-08-09)

 * Windows support
 * Uses Go context package (rather than golang.org/x/net/context)
 * Now requires Go 1.7

<hr/> 

## 1.0.0 (2016-10-13)

Initial release
 
