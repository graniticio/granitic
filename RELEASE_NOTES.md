**1.1.0  (2017-xx-xx)**

RDBMS:

 * Renamed types named RDBMSxxx to Rdbmsxxx to meet Go naming conventions
 * Streamlined [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms) methods (xxxTagxxx methods removed, xxxParamxxx methods more flexible with the type of arguments they accept.)
 * [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms) interface no longer requires DatabaseFromContext - implementations can now implement the optional ContextAwareDatabaseProvider interface.
 
Fixes:

 * [Issue 001](https://github.com/graniticio/granitic/issues/1) - Nilable fields on target objects now initialised if query binding in use
 * [Issue 002](https://github.com/graniticio/granitic/issues/2) - STOPALL validation operation now works as intended
 
<hr/> 

**1.0.1 (2017-08-09)**

 * Windows support
 * Uses Go context package (rather than golang.org/x/net/context)
 * Now requires Go 1.7

<hr/> 

**1.0.0 (2016-10-13)**

Initial release
 
