**1.1.0  (2017-xx-xx)**

RDBMS:

 * Streamlined RDBMSClient methods (xxxTagxxx methods removed, xxxParamxxx methods more flexible with the type of arguments they accept.)
 * DatabaseProvider interface no longer requires DatabaseFromContext
 
Fixes:

 * [Issue 001](https://github.com/graniticio/granitic/issues/1) - Nilable fields on target objects now initialised if query binding in use
 * [Issue 002](https://github.com/graniticio/granitic/issues/2) - STOPALL validation operation now works as intended
 
**1.0.1 (2017-08-09)**

 * Windows support
 * Uses Go context package (rather than golang.org/x/net/context)
 * Now requires Go 1.7
 

**1.0.0 (2016-10-13)**

Initial release
 
