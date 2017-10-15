# Tutorial - Reading data from an RDBMS

## What you'll learn

1. How to use MySQL and Docker to support this tutorial
1. How to connect Granitic to an RDBMS
1. How to template queries using Granitic's Query Manager facility
1. How to read data from an RDBMS into Go structs

## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Either have completed [tutorial 5](005-validation.md) or open a terminal and run:
 
<pre>
go get github.com/graniticio/granitic-examples
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 6
</pre>

## Setting up a test database

In order to streamline this tutorial, you will use an existing MySQL schema and test data supplied as part of the [Granitic Examples](https://github.com/graniticio/granitic-examples) 
GitHub repository in the <code>recordstore/graniticrs/db/schema-with-test-data.sql</code> file. 

If you haven't already cloned the [Granitic Examples](https://github.com/graniticio/granitic-examples) repo, you can do it with:

```
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
```

### Using an existing MySQL server

If you already have access to a MySQL database and are familiar with MySQL, you may run the <code>schema-with-test-data.sql</code> file against your existing database server and skip ahead to _Defining a DatabaseProvider_ below.
Note that the script creates a user <code>grnc</code> and allows it to connect from any IP.


### Docker and MySQL

[Install Docker](https://docs.docker.com/engine/installation/) then open a terminal and run:

```
cd $GOPATH/src/github.com/graniticio/granitic-examples/recordstore/graniticrs
docker-compose up --build -d
```

This will build and start a new docker image with the name <code>graniticrs_recordstore-db_1</code> and bind that image's port 3306 to your host machine's port 3306. 
If you want to stop the image you can run:

```json
docker stop graniticrs_recordstore-db_1
```

and destroy it permanently with

```json
docker rm graniticrs_recordstore-db_1
```

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

Go's SQL abstraction and 'driver management' models are much looser than some other language's RDBMS access layers. In
order to design Granitic's components and facilities so they are agnostic of the underlying RDBMS, an additional layer of abstraction
has been defined - the [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider).

The [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider) has two functions - to 
create instances of [sql.DB](https://golang.org/pkg/database/sql/#DB) (Go's connection/driver abstraction) and to define
a function that can recover the ID assigned to an inserted row (because not all DB drivers implement this in the same way).

Open a terminal and run

```json
go get -u github.com/go-sql-driver/mysql
```

to download the most widely used MySQL driver for Go then create a new file in your tutorial project <code>recordstore/db/provider.go</code> and set the contents to be:

```go
package db

import (
	"database/sql"
    "github.com/go-sql-driver/mysql"
	"context"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/rdbms"
)

type MySqlProvider struct {
	Config *mysql.Config
	Log logging.Logger
}

func (p *MySqlProvider) Database() (*sql.DB, error) {
	dsn := p.Config.FormatDSN()

	if db, err := sql.Open("mysql", dsn); err == nil {
		return db, nil
	} else {
		p.Log.LogErrorf("Unable to open connection to MySQL database: %v", err)

		return nil, err
	}
}

func (p *MySqlProvider) DatabaseFromContext(ctx context.Context) (*sql.DB, error) {
	return p.Database()
}

func (p *MySqlProvider) InsertIDFunc() rdbms.InsertWithReturnedID {
	return rdbms.DefaultInsertWithReturnedID
}
```

### New facilities and components

You'll need to enable two new facilities ([QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
and [RdbmsAccess](https://godoc.org/github.com/graniticio/granitic/facility/rdbms)) in your <code>resource/components/components.json</code> file:

```json
"Facilities": {
  "HttpServer": true,
  "JsonWs": true,
  "RuntimeCtl": true,
  "ServiceErrorManager": true,
  "QueryManager": true,
  "RdbmsAccess": true
}
```

In the same file, you'll need to declare a component for your [DatabaseProvider](https://godoc.org/github.com/graniticio/granitic/rdbms#DatabaseProvider) 
and a component to store your connection parameters:

```json
"dbProvider": {
  "type": "db.MySqlProvider"
},

"dbConnection": {
  "type": "mysql.Config",
  "User": "grnc",
  "Passwd": "OKnasd8!k",
  "Addr": "localhost",
  "DBName": "recordstore"
}
```

You've added components that rely on two new packages, so make sure you add:

```go
"github.com/go-sql-driver/mysql",
"granitic-tutorial/recordstore/db"
````

to the packages section at the start of <code>components.json</code> file.


### Configuration in components.json

Directly storing the database connection parameters in the <code>components.json</code> file is bad practise and is only used here
to keep the length of this tutorial down. Instead refer back to [the configuration tutorial](002-configuration.md) to see how
you could use config promises and a separate configuration file to store this type of environment-specific configuration.


## Artist GET

We are going to connect our existing <code>/artist GET</code> endpoint to the database. Modify the <code>endpoint/artist.go</code> 
file so that the <code>ArtistLogic</code> type looks like:

```go

```  







## Building a query template

An optional (but recommended) facility offered by Granitic is the [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager).
This facility allows you to define your database queries in text files outside of your Go code and have variables injected into
the template at runtime to create a working query.

The [QueryManager](https://godoc.org/github.com/graniticio/granitic/facility/querymanager) 
uses <code>resource/queries</code> as the default location for templates, so create a new file <code>recordstore/resource/queries/search</code> 

## Start and test

At this point your service can be started:

```go
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind && go build && ./recordstore -c resource/config
``` 

and visiting <code>http://localhost:8080/artist-search</code> will yield a response object containing an empty array