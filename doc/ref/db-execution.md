# Query execution

---

## Obtaining a Client interface

Queries can be executed against an RDBMS by any code that has access to a [client manager](db-provider.md). Typically your code
will be part of a struct which has a member variable of type [rdbms.ClientManager](https://godoc.org/github.com/graniticio/granitic/rdbms#ClientManager)
into which Granitic will inject a client manager.

```go
type ArtistDAO struct {
    DBClientManager rdbms.ClientManager
} 
```

A method in your application can then use this manager to obtain a [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/rdbms#Client),
which is the interface that allows you to execute queries:

```go
func (ad *ArtistDAO) FindAll() ([]*Artist,error) {

    client, err := ad.DBClientManager.Client()

    if err != nil {
        return nil, err
    }
}
```

or 

```go
func (ad *ArtistDAO) FindAll(ctx context.Context) ([]*Artist,error) {

    client, err := ad.DBClientManager.ClientFromContext(ctx)

    if err != nil {
        return nil, err
    }
}
```

Note that a [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/rdbms#Client) is **not** goroutine safe as
it holds information about transaction state and implicitly about a connection to a database.

## Executing templated queries

The [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/rdbms#Client) interface is designed to work
closely with the [query manager](db-query.md) and provides a series of methods to execute templated queries.
Those methods are named in a similar way to those on [regexp.Regexp](https://golang.org/pkg/regexp/#Regexp) and are 
generally of the form:

```
  SQLVerb[BindingType]QID[ParameterSource]
```

Where

  * SQLVerb is Select, Delete, Update or Insert
  * BindingType is optional and can be Bind or BindSingle
  * ParameterSource is optional and can be either Param or Params

QID indicates the method is expecting to be passed the ID of a [query template](db-query.md) managed by the [query manager](db-query.md)

### Binding 

Methods with `Bind` or `BindSingle` provide a mechanism for automatically copying result data into structs or slices of 
structs. 

#### BindSingle

If the method name contains `BindSingle`, you will pass a pointer to a struct into the method and its fields will be populated:

```go
ad := new(ArtistDetail)

if found, err := rc.SelectBindSingleQIDParams("ARTIST_DETAIL", rid, ad); found {
  return ad, err
} else {
  return nil, err
}
```

This is typically used for queries that are guaranteed to return a single row.

#### Bind

If the method contains the word `Bind`, you will supply an example 'template' instance of a struct and 
the method will return a slice of that type:

```go
params := make(map[string]interface{}

if r, err := client.SelectBindQIDParams("ARTIST_SEARCH_BASE", params, new(ArtistSearchResult)); err != nil {
  return nil, err
} else {
  return id.artistResults(r), nil
}
```

This is typically used for any query which returns an unknown number of rows.

### Parameters sources

Parameters to populate template queries can either be supplied via a single name/value pair (methods with the word `Param`)
or via a map or struct (methods with `Params`).

When using a map, it should be of type `map[string]interface{}`. When using structs, the field names on the struct
will be matched against expected variable names, unless the `dbparam` struct tag is used instead.

See the [Variables section](db-query.md) of the query manager documentation for more details.

## Executing non-templated queries

If you are not using the [query manager](db-query.md), [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/rdbms#Client)
provides pass through access to the `Exec`, `Query` and `QueryRow` methods on [sql.DB](https://golang.org/pkg/database/sql/#DB)

Not that these direct methods can be freely mixed with calls to the templated-query methods, even inside a transaction
(see below).


## Transactions

To start a transaction, invoke the `StartTransaction` method on the `Client` like:

```go
client.StartTransaction()
defer db.Rollback()
````

and end it with:

```go
client.CommitTransaction()
```

The deferred `Rollback` call will do nothing if the transaction has previously been committed.

Any query methods executed between `StartTransaction` and `CommitTransaction`/`Rollback` will be executed
inside a 'transaction' as defined by your RDBMS.

## Utility methods

The [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/rdbms#Client) interface provides a number of utility
methods to support templated query execution.

`FindFragment` returns the text of a query template that does not have any variables.

`RegisterTempQuery` registers a query (not a template) with the given ID, allow those methods that expect a QID
to be used with a query not managed by the query manager.

These two methods can be used to support the construction of (some) dynamic queries without embedding SQL in your
application's Go code.


---
**Next**: [Scheduled activities](sch-index.md)

**Prev**: [Query management](db-query.md)

