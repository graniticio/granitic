# Tutorial - Writing data to a database

## What you'll learn

1. How to insert data into your database and capture generated IDs
1. How to make your database calls transactional

## Prerequisites

 1. Follow the Granitic [installation instructions](../installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Follow the [setting up a test database](006-database-read.md) section of [tutorial 6](006-database-read.md)
 1. Either have completed [tutorial 6](006-database-read.md) or clone the
  [tutorial repo](https://github.com/graniticio/tutorial) and navigate to `json/007/recordstore` in your terminal.
 

## Test database

If you didn't complete [tutorial 6](006-database-read.md), please return to it and work through the 
'[Setting up a test database](006-database-read.md)' section which explains how to run Docker and MySQL with a 
pre-built test database.


## Inserting data

Our tutorial application already allows web service clients to submit a new artist to be stored in our database using the 
`/artist POST` endpoint, but it currently just simulates an insert. To alter this code to actually store data, 
open the `resource/queries/artist` file in your tutorial project and add the following query:

```sql
ID:CREATE_ARTIST

INSERT INTO artist(
  name,
  first_active
) VALUES (
  ${!Name},
  ${FirstYearActive}
)
```

You'll notice that the names of the variables match the field names on the `artist.Submission` struct in 
`artist/post.go`. You'll also notice we're not inserting an ID for this new record. The `artist`
table on our test database _does_ have an ID column defined as:

```sql
  id INT NOT NULL AUTO_INCREMENT
```

so a new ID will be generated automatically. We'll show you how to capture that ID shortly. 

### Required parameters

You might have noticed that the `Name` parameter is referenced in the template as `${!Name}`. The exclamation mark
indicates that the parameter is required and the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
will return an error if this parameter is missing.

As the `FirstYearActive` parameter maps to the nullable column:

```sql
  artist.first_active SMALLINT
```

in the database, it is _not_ marked as required. If the `FirstYearActive` parameter is missing (or set to `nil`),
the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) will substitute the value 
`null` when generating the query, because we configured the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
to run in `SQL` mode in the previous tutorial.


## Executing the query and capturing the ID

Modify the `PostLogic` struct in `artist/post.go` so it looks like:

```go
type PostLogic struct {
  Log logging.Logger
  DbClientManager rdbms.ClientManager
}

func (pl *PostLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, s *Submission) {
  // Obtain an RdmsClient from the rdbms.RdbmsClientManager injected into this component
  dbc, _ := pl.DbClientManager.Client()

  // Declare a variable to capture the ID of the newly inserted artist
  var id int64

  // Execute the insert, storing the generated ID in our variable
  if err := dbc.InsertCaptureQIDParams("CREATE_ARTIST", &id, s); err != nil {
    // Something went wrong when communicating with the database - return HTTP 500
    pl.Log.LogErrorf(err.Error())
    res.HTTPStatus = http.StatusInternalServerError
  }

  // Use the new ID as the HTTP response, wrapped in a struct
  res.Body = CreatedResource{id}
}
```

The import section of this file should now look something like:

```go
import (
  "context"
  "github.com/graniticio/granitic/v3/logging"
  "github.com/graniticio/granitic/v3/rdbms"
  "github.com/graniticio/granitic/v3/types"
  "github.com/graniticio/granitic/v3/ws"
  "net/http"
)
```

## Building and testing

Start your service by opening a terminal, navigating to your tutorial project folder and running:

```
grnc-bind && go build && ./recordstore
```

and POST the following JSON to `http://localhost:8080/artist`

```json
{
  "Name": "Another Artist",
  "FirstYearActive": 2010
}
```

(see the [data capture tutorial](004-data-capture.md) for instructions on using a browser plugin to do this)

You should see a response like:

```json
{
  "ID": 10
}
```

and the ID will increment by one each time you re-POST the data as new rows are inserted into the database.


## Transactions

Only the simplest web service endpoints insert a single row - more often than not you'll need to make multiple
inserts and updates to create a resource. In this case you'll need to make your database calls as part of a transaction. 
We will illustrate this by allowing the web service caller to provide a list of 'related artists' when calling `/artist POST`. 

The test database contains a table that stores this relationship:

```sql
CREATE TABLE related_artist (
  artist_id INT NOT NULL,
  related_artist_id INT NOT NULL,
  FOREIGN KEY (artist_id) REFERENCES artist(id),
  FOREIGN KEY (related_artist_id) REFERENCES artist(id),
  UNIQUE (artist_id, related_artist_id)
)
```

First alter the definition of the `artist.Submission` so it looks like:

```go
type Submission struct {
  Name            *types.NilableString
  FirstYearActive *types.NilableInt64
  RelatedArtists []int64
}
```

Create a new query in your `resource/queries/artist` file:

```sql
ID:RELATE_ARTIST

INSERT INTO related_artist(
  artist_id,
  related_artist_id
) VALUES (
  ${!ArtistID},
  ${!RelatedArtistID}
)
```


And modify the `PostLogic.ProcessPayload` method so it looks like:


```go
func (pl *PostLogic) ProcessPayload(ctx context.Context, req *ws.Request, res *ws.Response, s *Submission) {
  // Obtain a Client from the rdbms.ClientManager injected into this component
  dbc, _ := pl.DbClientManager.Client()

  defer dbc.Rollback()

  // Start a database transaction
  dbc.StartTransaction()

  // Declare a variable to capture the ID of the newly inserted artist
  var id int64

  // Execute the insert, storing the generated ID in our variable
  if err := dbc.InsertCaptureQIDParams("CREATE_ARTIST", &id, s); err != nil {
    // Something went wrong when communicating with the database - return HTTP 500
    pl.Log.LogErrorf(err.Error())
    res.HTTPStatus = http.StatusInternalServerError
  }

  // Insert a row for each related artist
  params := make(map[string]interface{})
  params["ArtistID"] = id

  for _, raID := range s.RelatedArtists {
    params["RelatedArtistID"] = raID

    if _, err := dbc.InsertQIDParams("RELATE_ARTIST", params); err != nil {
      // Something went wrong inserting the relationship
      pl.Log.LogErrorf(err.Error())
      res.HTTPStatus = http.StatusInternalServerError

      return
    }

  }

  // Commit the transaction
  dbc.CommitTransaction()

  // Use the new ID as the HTTP response, wrapped in a struct
  res.Body = CreatedResource{id}
}
```

A few things are worth noting here:

  1. When using transactions, it's good practice to call `defer dbc.Rollback()` as this means that the transaction 
  will be explicitly rolled back if we return from the function or if there is a panic.
  1. Calling `Rollback` has no effect if the tranasction has already been commited.
  1. `StartTransaction`, `Rollback` and `Commit` can all return errors which are being ignored in
  this example.
  
Rebuild and restart your new service and test by posting:

```json
{
  "Name": "Artist with friends",
  "RelatedArtists":[1,2]
}
```

## Including database checks as part of validation

At moment our code will fail with a database error if we supply a 'related artist' ID that isn't actually the ID of an artist
in the database. Try sending:

```json
{
  "Name": "Artist with friends",
  "RelatedArtists":[-1000]
}
```

and you'll get a generic Granitic HTTP 500 error and the following in the logs:

```
ERROR submitArtistLogic Error 1452: Cannot add or update a child row: a foreign key constraint fails
```

Although the integrity of the data in the database has been protected, this isn't very elegant. What we'd like is to
have the submitted IDs checked during validation. We can do this by creating a component that implements the [validate.ExternalInt64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalInt64Validator)
interface.

The [next tutorial](008-shared-validation.md) explains how to do this, but if you haven't followed the [validation tutorial](005-validation.md) now would be a good time to familiarise yourself with it.


## Recap
  * Parameters in query templates can be marked as required by prefixing the parameter name with !
  * [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClient) provides the methods for 
    managing transactions
  * It is good practice to `defer` a call to `Rollback()` to make sure transactions are explicitly rolled back if 
    something goes wrong.
  
## Further reading

 * [RDBMS GoDoc](https://godoc.org/github.com/graniticio/granitic/rdbms)
 * [QueryManager GoDoc](https://godoc.org/github.com/graniticio/granitic/facility/querymanager)
 
## Next

The next tutorial covers [sharing validation rules](008-shared-validation.md) between endpoints and using custom 
Granitic components to validate fields. 