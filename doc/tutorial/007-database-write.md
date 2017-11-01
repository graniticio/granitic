# Tutorial - Writing data to a database

## What you'll learn

1. How to insert data into your database and capture generated IDs
1. How to make your database calls transactional

## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Either have completed [tutorial 6](006-database-read.md) or open a terminal and run:
 1. Followed the [setting up a test database](006-database-read.md) section of [tutorial 6](006-database-read.md)
 
<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 7
</pre>


## Test database

If you didn't follow the [tutorial 6](006-database-read.md), please work through the '[Setting up a test database](006-database-read.md)'
section which explains how to run Docker and MySQL with a pre-built test database.


## Inserting data

Our tutorial already allows web service clients to submit a new artist to be stored in our record store database using the 
<code>/artist POST</code> endpoint, but it currently just simulates an insert. To alter this code to actually store data, 
open the <code>resource/queries/artist</code> file and add the following query:

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

You'll notice that the names of the variables match the field names on the <code>SubmittedArtistRequest</code> struct in 
<code>endpoint/artist.go</code>. You'll also notice we're not inserting an ID for this new record. The <code>artist</code>
table on our test database _does_ have an ID column defined as:

```sql
  id INT NOT NULL AUTO_INCREMENT
```

so a new ID will be generated automatically. We'll show you how to capture that ID shortly. 

### Required parameters

You might have noticed that the <code>Name</code> parameter is referenced in the template as <code>${!Name}</code>. The exclamation mark
indicates that the parameter is required and the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
will return an error if this parameter is missing.

As the <code>FirstYearActive</code> parameter maps to the nullable column:

```sql
  artist.first_active SMALLINT
```

in the database, it is not marked as required. If the <code>FirstYearActive</code> parameter is missing (or set to <code>nil</code>),
the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) will substitute the value <code>null</code> when generating 
the query, because we configured the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) to run in <code>SQL</code> 
mode in the previous.


## Executing the query and capturing the ID

Modify the <code>SubmitArtistLogic</code> struct in <code>endpoint/artist.go</code> so it looks like:

```go
type SubmitArtistLogic struct {
  Log logging.Logger
  DbClientManager rdbms.RdbmsClientManager
}

func (sal *SubmitArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

  sar := req.RequestBody.(*SubmittedArtistRequest)

  // Obtain an RdmsClient from the rdbms.RdbmsClientManager injected into this component
  dbc, _ := sal.DbClientManager.Client()

  // Declare a variable to capture the ID of the newly inserted artist
  var id int64

  // Execute the insert, storing the generated ID in our variable
  if err := dbc.InsertCaptureQIdParams("CREATE_ARTIST", &id, sar); err != nil {
    // Something went wrong when communicating with the database - return HTTP 500
    sal.Log.LogErrorf(err.Error())
    res.HttpStatus = http.StatusInternalServerError
  }

  // Use the new ID as the HTTP response, wrapped in a struct
  res.Body = struct {
    Id int64
  }{id}

}

func (sal *SubmitArtistLogic) UnmarshallTarget() interface{} {
  return new(SubmittedArtistRequest)
}
```

## Building and testing

Start your service:

```
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind && go build && ./recordstore -c resource/config
```

and POST the following JSON to <code>http://localhost:8080/artist</code>

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
  "Response": {
    "Id": 10
  }
}
```

and the ID will increment by one each time you re-POST the data.


## Transactions

Only the simplest web service endpoints insert a single row - more often than not you'll need to make multiple
inserts and updates. In most cases, you'll need to make your database calls as part of a transaction. We will illustrate this by
allowing the web service caller to provide a list of 'related artists' when calling <code>/artist POST</code>. 

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

First alter the definition of the <code>SubmittedArtistRequest</code> so it looks like:

```go
type SubmittedArtistRequest struct {
  Name            *types.NilableString
  FirstYearActive *types.NilableInt64
  RelatedArtists []int64
}

```

Create a new query in your <code>resource/queries/artist</code> file:

```sql

ID:RELATE_ARTIST

INSERT INTO related_artist(
  artist_id,
  related_artist_id
) VALUES (
  ${!ArtistId},
  ${!RelatedArtistId}
)
```


And modify the <code>SubmitArtistLogic.Process</code> method so it looks like:


```go
func (sal *SubmitArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

	sar := req.RequestBody.(*SubmittedArtistRequest)

	// Obtain an RdmsClient from the rdbms.RdbmsClientManager injected into this component
	dbc, _ := sal.DbClientManager.Client()
	defer dbc.Rollback()

	// Start a database transaction
	dbc.StartTransaction()

	// Declare a variable to capture the ID of the newly inserted artist
	var id int64

	// Execute the insert, storing the generated ID in our variable
	if err := dbc.InsertCaptureQIdParams("CREATE_ARTIST", &id, sar); err != nil {
		// Something went wrong when communicating with the database - return HTTP 500
		sal.Log.LogErrorf(err.Error())
		res.HttpStatus = http.StatusInternalServerError

		return

	}

	// Insert a row for each related artist
	params := make(map[string]interface{})
	params["ArtistId"] = id

	for _, raId := range sar.RelatedArtists {
		params["RelatedArtistId"] = raId

		if _, err := dbc.InsertQIdParams("RELATE_ARTIST", params); err != nil {
			// Something went wrong inserting the relationship
			sal.Log.LogErrorf(err.Error())
			res.HttpStatus = http.StatusInternalServerError

			return
		}

	}

	// Commit the transaction
	dbc.CommitTransaction()

	// Use the new ID as the HTTP response, wrapped in a struct
	res.Body = struct {
		Id int64
	}{id}

}
```

A few things are worth noting here:

  1. When using transactions, it's good practice to call <code>defer dbc.Rollback()</code> this means that the transaction 
  will be explicitly rolled back if we return from the function or if there is a panic.
  1. Calling <code>Rollback</code> has no effect if the tranasction has already been commited.
  1. <code>StartTransaction</code>, <code>Rollback</code> and <code>Commit</code> can all return errors which are being ignored in
  this example.
  
Rebuild and restart your new service and test it with:

```json
{
  "Name": "Artist with friends",
  "RelatedArtists":[1,2]
}
```

## Recap

 * Granitic requires your code to implement a [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider)  component.
 * [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClient) objects provide the interface for executing queries.
 * These are obtained through the [RdbmsClientManager](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClientManager) framework component which is automatically
 injected into your application components if they have a field *DbClientManager rdbms.RdbmsClientManager*
 * Queries can be stored in template files and accesed by your code using IDs. This feature is provided by the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) facility.
 
## Further reading

 * [RDBMS GoDoc](https://godoc.org/github.com/graniticio/granitic/rdbms)
 * [QueryManager GoDoc](https://godoc.org/github.com/graniticio/granitic/facility/querymanager)
 
## Next

The next tutorial covers the [writing of data to an RDBMS](007-database-write.md)