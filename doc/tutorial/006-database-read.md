# Tutorial - Reading data from an RDBMS

## What you'll learn

1. How to use MySQL and Docker to support this tutorial
1. How to connect Granitic to an RDBMS
1. How to build templates for your SQL queries using Granitic's QueryManager facility
1. How to read data from an RDBMS into Go structs

## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/v2/blob/master/doc/installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Either have completed [tutorial 5](005-validation.md) or open a terminal and run:
 
<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 6
</pre>

## Setting up a test database

In order to streamline this tutorial, you will use an existing MySQL schema and test data supplied as part of the [Granitic Examples](https://github.com/graniticio/granitic-examples) 
GitHub repository in the <code>tutorial/docker/db/schema-with-test-data.sql</code> file. 

If you haven't already cloned the [Granitic Examples](https://github.com/graniticio/granitic-examples) repo, you can do it with:

<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
</pre>

### Using an existing MySQL server

If you already have access to a MySQL database and are familiar with MySQL, you may run the <code>schema-with-test-data.sql</code> file against your existing database server and skip ahead to _Defining a DatabaseProvider_ below.
Note that the script creates a user <code>grnc</code> and allows it to connect from any IP.


### Docker and MySQL

[Install Docker](https://docs.docker.com/engine/installation/) then open a terminal and run:

<pre>
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial/docker
docker-compose up --build -d
</pre>

This will build and start a new docker image with the name <code>graniticrs_recordstore-db_1</code> and bind that image's port 3306 to your host machine's port 3306. 
If you want to stop the image you can run:

<pre>
docker stop graniticrs_recordstore-db_1
</pre>

and destroy it permanently with

<pre>
docker rm graniticrs_recordstore-db_1
</pre>

### MySQL workbench

If you are interested in examining the database's structure and/or data, you can connect to the new database with [MySQL Workbench](https://dev.mysql.com/downloads/workbench/)
(or any other tool) using the following credentials:

```
Host:     localhost 
Port:     3306
User:     grnc
Password: OKnasd8!k
Schema:   recordstore
```


## Creating a DatabaseProvider

Go's SQL abstraction and 'driver management' models are much looser than some other languages' RDBMS access layers. In
order to allow Granitic's components and facilities to be agnostic of the underlying RDBMS, an additional layer of abstraction
has been defined - the [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider) interface.
Your application will have to define a component that implements this interface. 

The [DatabaseProvider's](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider) 
role is to create instances of [sql.DB](https://golang.org/pkg/database/sql/#DB) (Go's connection/driver abstraction) and 
implement any connection pooling and load balancing your application requires. It's also the most convinient place to import 
whichever package provides the database driver that you require


### Obtaining a MySQL driver for Go

Open a terminal and run

```json
go get -u github.com/go-sql-driver/mysql
```

to download the most widely used MySQL driver for Go.

### Creating a DatabaseProvider component

Create a new file in your tutorial project <code>recordstore/db/provider.go</code> and set the contents to be:

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
}

func (p *MySQLProvider) Database() (*sql.DB, error) {
  dsn := p.Config.FormatDSN()
  
  if db, err := sql.Open("mysql", dsn); err == nil {
    return db, nil
  } else {
    p.Log.LogErrorf("Unable to open connection to MySQL database: %v", err)
    
    return nil, err
  }
}

```

In this file we are importing the database driver package and providing a method for Granitic to call when it needs a connection
to your database. The implementation here is very simple and doesn't offer any connection management other than that implemented
by the driver itself.

In your <code>resource/components/components.json</code> file you'll need to declare a component for your [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider) 
and a component to store your connection parameters:

```json
"dbProvider": {
  "type": "db.MySQLProvider",
  "Config": "ref:dbConnection"
},

"dbConnection": {
  "type": "mysql.Config",
  "User": "grnc",
  "Passwd": "OKnasd8!k",
  "Addr": "localhost",
  "DBName": "recordstore",
  "AllowNativePasswords": true
}
```

You've added components that rely on two new packages, so make sure you add:

```go
"github.com/go-sql-driver/mysql",
"granitic-tutorial/recordstore/db"
```

to the packages section at the start of <code>components.json</code>


### Configuration in components.json

Directly storing the database connection parameters in the <code>components.json</code> file is bad practice and is only used here
to keep the length of this tutorial down. Refer back to [the configuration tutorial](002-configuration.md) to see how
you could use config promises and a separate configuration file to store this type of environment-specific configuration.


## New facilities

You'll need to enable two new facilities ([QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
and [RdbmsAccess](https://godoc.org/github.com/graniticio/granitic/facility/rdbms)) in your <code>resource/config/config.json</code>

```json
"Facilities": {
  "HTTPServer": true,
  "JSONWs": true,
  "RuntimeCtl": true,
  "ServiceErrorManager": true,
  "QueryManager": true,
  "RdbmsAccess": true
}
```


### RdbmsAccess

This facility is the bridge between Granitic's database framework and your application code. It uses the [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider)
to obtain connections to your database and injects an instance of [RdbmsClientManager](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClientManager) into
any of your application components that have the field:

```go
  DbClientManager rdbms.RdbmsClientManager
```

### QueryManager

An optional (but recommended) facility offered by Granitic is the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager).
This facility allows you to define your database queries in text files outside of your Go code and have variables injected into the template at runtime to create a working query.

The [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) facility is not intended to be specific
to relational databases; it is designed to support any data source that supports a query language (e.g. search engines, NoSQL databases).

However, it can be configured to provide additional support for SQL queries, so add this to your <code>config.json</code> file:

```json
"QueryManager": {
  "ProcessorName": "sql"
 }
```

## Artist GET

We are going to connect our existing <code>/artist GET</code> endpoint to the database. Modify the <code>endpoint/artist.go</code> 
file so that the <code>ArtistLogic</code> type looks like:

```go
type ArtistLogic struct {
  EnvLabel string
  Log      logging.Logger
  DbClientManager rdbms.RdbmsClientManager
}

func (al *ArtistLogic) ProcessPayload(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse, ar *ArtistRequest) {
  // Obtain an RdmsClient from the rdbms.RdbmsClientManager injected into this component
  dbc, _ := al.DbClientManager.Client()

  // Create a new object to store the results of our database call 
  result := new(ArtistDetail)

  // Call the database and populate our object
  if found, err := dbc.SelectBindSingleQIDParams("ARTIST_BY_ID", result, ar); found {
    // Make our result object the body of the HTTP response we'll send
    res.Body = result
  
  } else if err != nil{
    // Something went wrong when communicating with the database - return HTTP 500
    al.Log.LogErrorf(err.Error())
    res.HTTPStatus = http.StatusInternalServerError

  } else {
    // No results were returned by the database call - return HTTP 404
    res.HTTPStatus = http.StatusNotFound
  }
}

```  

The imports section of this file should now be:

```go
import (
  "context"
  "github.com/graniticio/granitic/v2/logging"
  "github.com/graniticio/granitic/v2/types"
  "github.com/graniticio/granitic/v2/ws"
  "github.com/graniticio/granitic/v2/rdbms"
  "net/http"
)
```

### RdbmsClient

[RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClient) is the interface your code uses to execute queries and manage transasctions. It is *not* goroutine-safe and should not be shared, which is why
we use the <code>RdbmsClientManager</code> to create a new instance on every request. 

The methods on [RdbmsClient](https://godoc.org/github.com/graniticio/granitic/rdbms#RdbmsClient) are named to make the intent of your database calls more obvious, in this case the method <code>SelectBindSingleQIDParams</code> tells us:

 * Select - You are executing a SELECT-type SQL query
 * BindSingle - You expect zero or one results and want the result bound into a supplied object (the *ArtistDetail*)
 * QID - You are supplying the Query ID of a previously templated query to execute
 * Params - You are supplying one or more objects that can be used to inject values into your templated queries (the *ArtistRequest*)

There are a number of variations on these methods, including binding multi-row queries into a slice of objects of your choice. Refer to the [rdbms GoDoc](https://godoc.org/github.com/graniticio/granitic/rdbms)
for more information.


## Building a query template

The [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
uses <code>resource/queries</code> as the default location for templates, so create a new file <code>recordstore/resource/queries/artist</code> 
and set the contents to:

```sql
ID:ARTIST_BY_ID

SELECT
  name AS Name
FROM
  artist
WHERE
  id = ${ID}
  
```  

Each file can contain any number of queries. The line starting <code>ID:</code> delimits the queries and assigns an ID to
the following query (in this case <code>ARTIST_BY_ID</code>). Variables are surrounded by <code>${}</code> and names are case sensitive.


### Parameters to queries

Values for parameters are injected into the query template when you call a method on <code>RdbmsClient</code>. In the case of this query the <code>${ID}</code> parameter will be populated when we call:

```go
  dbc.SelectBindSingleQIDParams("ARTIST_BY_ID", result, ar)
```

because the <code>ArtistRequest</code> object we are passing as the 'parameter source' has a field named ID. If you want to use a different parameter name in 
your query, you can use the <code>dbparam</code> struct tag like:

```go
type ArtistRequest struct {
  ID  int `dbparam:"artist-id"`
}
```

or you can supply a <code>map[string]interface{}</code> as a source of parameters instead of a struct and have complete control over the
names of the map keys.

### Matching column names and aliases

Granitic automatically populates the fields on the supplied target object (in this case an instance of <code>ArtistDetail</code>) by finding a field 
whose name and type matches a column name or alias in the SQL query's results. This process is case sensitive, which is why we've had to define
the column alias <code>Name</code> in the query to match it to the <code>ArtistDetail.Name</code> field.

If you'd prefer not to use aliases in your query or want the the name of the field on the struct to be very different to the column name, you
can use the <code>column</code> tag on your target object, e.g.:

```go
type ArtistDetail struct {
 	Name string `column:name`
} 
```

#### Debugging queries

You can make the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) log the queries it constructs by
setting the <code>grncQueryManager</code> framework component's log level to <code>DEBUG</code> in your config file:

```json
"FrameworkLogger": {
  "GlobalLogLevel": "INFO",
  "ComponentLogLevels": {
    "grncQueryManager": "DEBUG"
  }
}
``` 

or at runtime from your command line:

```
  grnc-ctl log-level grncQueryManager DEBUG
```

Refer to the [logging tutorial](003-logging.md) for more information on how this works.


### Query ID definition

By default, query IDs are defined in your template file with a line like:

```
ID:YOUR_QUERY_ID
```

If your query files only contain SQL queries, you'll probably want edit them in an IDE or editor that has SQL syntax highlighting
and checking. Your editor will complain about the lines where query IDs are defined.

If you add the following to your config:

```json
"QueryManager": {
  "QueryIDPrefix": "--ID:"
}
``` 

you can define your query IDs like:

```
--ID:YOUR_QUERY_ID
```

and your editor will stop complaining.

## Start and test

At this point your service can be started:

```
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind && go build && ./recordstore -c resource/config
``` 

and visiting <code>http://localhost:8080/artist/1</code> will yield a response like:

```json
{
  "Name":"Younger artist"
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