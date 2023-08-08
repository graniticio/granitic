# Database providers and client managers

[Reference](README.md) | [Relational Databases](db-index.md)

---

## Database provider

A database provider is the name Granitic gives to a component in your application which:

  * Is responsible for importing the correct database driver into your application
  * Allows your code to implement advanced connection management if needed
  * Allows your code to provide specific alternative to some driver specific behaviour
  
You need a database provider if you want to connect to a relational (SQL) database from your Granitic application.
  
### Minimal implementation

For the most common use case where your application will connect to a single instance of a database 
you must provide a single component which implements [rdbms.DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#DatabaseProvider):

```go
type DatabaseProvider interface {
  Database() (*sql.DB, error)
}
```
 
A minimal implementation (for connecting to a MySQL database), might look like:

```go
package db

import (
  "database/sql"
  "github.com/go-sql-driver/mysql"
  "github.com/graniticio/granitic/v2/logging"
)

type MySQLProvider struct {
  Config *mysql.Config
  Log logging.Logger
  db *sql.DB
}

func (p *MySQLProvider) Database() (*sql.DB, error) {
  dsn := p.Config.FormatDSN()
  
  if p.db != nil {
    return p.db, nil 
  }

  if db, err := sql.Open("mysql", dsn); err == nil {
    return db, nil
  } else {
    p.Log.LogErrorf("Unable to open connection to MySQL database: %v", err)
    
    return nil, err
  }
}
```

with a corresponding component definition:

```json
"dbProvider": {
  "type": "db.MySQLProvider",
  "Config": {
    "type": "mysql.Config",
    "User": "$database.user",
    "Passwd": "$database.password",
    "Addr": "$database.host",
    "DBName": "$database.instance",
    "AllowNativePasswords": true
  }
} 
```

### Caching the DB object

It is important to note that the [*sql.DB](https://golang.org/pkg/database/sql/#DB) object returned by your provider is a 
'database handle representing a  pool of zero or more underlying connections. It's safe for concurrent use by multiple goroutines'.

Generally your `DatabaseProvider` should only create a DB instance once, store it in a member variable and then
return the cached `*sql.DB` on subsequent requests (as in the above example).

### Context aware provider

Your provider may optionally implement [rdbms.ContextAwareDatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#ContextAwareDatabaseProvider):

```go
type ContextAwareDatabaseProvider interface {
  DatabaseFromContext(context.Context) (*sql.DB, error)
}
```

Having access to the context allows your provider implementation to alter its behaviour based on the contents of the 
context. For example, you might use a different database shard
depending on some user ID stored in the context.

### Blocking startup until connected

Refer to the [RdbmsAccess facility](fac-rdbms.md) documentation on how to delay application startup until a database connection 
has been established.

### IDs on insert

A common pattern is for a SQL insert query to rely on a database specific mechanism (an auto-increment or a sequence)
to create a new numeric primary key (or unique ID) when inserting a row. It is often vital for your application code to know
what that new ID is, but the implementation for recovering that ID is not standardised across Go SQL drivers.

By default Granitic uses the function [rdbms.DefaultInsertWithReturnedID](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#DefaultInsertWithReturnedID)
to return the newly created ID (which relies on the `LastInsertId` function on [sql.Result](https://golang.org/pkg/database/sql/#Result`)).

If this doesn't work for your RDBMS, your database provider should implement [rdbms.NonStandardInsertProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#NonStandardInsertProvider):

```go
type NonStandardInsertProvider interface {
  InsertIDFunc() InsertWithReturnedID
}
```

## Client managers

A client manager is a Granitic system component that is injected automatically into any component of yours that has a field:

```go
DBClientManager rdbms.ClientManager
``` 

once you enable the [RdbmsAccess facility](fac-rdbms.md) and have a component that implements 
[rdbms.DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#DatabaseProvider)

At runtime, your code can call the `Client` or `ClientFromContext` method on the client manager to obtain an
[rbms.Client](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#Client) which is the (non-goroutine safe)
interface that your code will use to execute queries, the usage of which is covered in [executing queries](db-query.md).

## Working with multiple databases

If your application connects to more than one database, configuration is more complex. For each database you want to 
connect to, you will need:

 * A component implementing [rdbms.DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#DatabaseProvider)
 * A component that is an instance of [rdbms.ClientManagerConfig](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#ClientManagerConfig)

The `ClientManagerConfig` component defines:

 * A reference to the database provider to use
 * The names of fields on your components that are of type `ClientManager` and will receive this ClientManager
 * A logical name for the client manager
 
### Example

```json
"catalogueDBProvider": {
  "type": "db.MySQLProvider",
  "Config": {
    "type": "mysql.Config",
    "User": "$database.catalogue.user",
    "Passwd": "$database.catalogue.password",
    "Addr": "$database.catalogue.host",
    "DBName": "$database.catalogue.instance",
    "AllowNativePasswords": true
  }
}, 

"ordersDBProvider": {
  "type": "db.MySQLProvider",
  "Config": {
    "type": "mysql.Config",
    "User": "$database.orders.user",
    "Passwd": "$database.orders.password",
    "Addr": "$database.orders.host",
    "DBName": "$database.orders.instance",
    "AllowNativePasswords": true
  }
}, 

"catalogueManager": {
  "type": "rdbms.ClientManagerConfig",
  "Provider": "+catalogueDBProvider",
  "InjectFieldNames": ["CatalogueClientManager"],
  "BlockUntilConnected": true
},

"ordersManager": {
  "type": "rdbms.ClientManagerConfig",
  "Provider": "+ordersDBProvider",
  "InjectFieldNames": ["OrdersClientManager", "Orders"],
  "BlockUntilConnected": true
}
``` 

This example shows an application that connects to two different databases. Because they are the same type of database
(MySQL in this instance), the providers can re-use the same underlying implementation.

Two different client managers will be created. Each has a reference to a specific database provider and defines the name(s)
of the fields that the managers will be auto-injected into (as long as the field is of type `ClientManager`).

Both client managers are configured to block application startup until a successful database connection is established.




---
**Next**: [Query management](db-query.md)

**Prev**: [Principles](db-principles.md)