# Release notes

## 1.1.0  (2017-xx-xx)

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
 
### HttpServer

 * Contexts passed to WsHandler are now inherited from the context on http.Request
 * Server makes use of new Shutdown and Close http.Server methods during framework shutdown

 
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
 
