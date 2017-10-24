# Tutorial - Reading data from an RDBMS

## What you'll learn

1. How to inset data and capture generated IDs
1. How to make your database calls transactional
1. How to add database calls to your automatic validation

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


## Inserting data and capturing an ID

Our tutorial already allows web service clients to submit a new artist to be stored in our record store database using the 
<code>/artist POST</code> endpoint, but it currently just simulates an insert. To alter this code to actually store data, 
open the <code>resource/queries/artist</code> file and add the following:

```mysql
ID:CREATE_ARTIST

INSERT INTO artist(
  name,
  firstActive
) VALUES (
  ${Name},
  ${FirstYearActive}
)
```

You'll notice that the names of the variables match the field names on the <code>SubmittedArtistRequest</code> struct in 
<code>endpoint/artist.go</code>. You'll also notice we're not inserting an ID for this new record. The <code>artist</code>
table on our test database _does_ have an ID column defined as:

```mysql
  id INT NOT NULL AUTO_INCREMENT
```

so a new ID will be generated automatically. We'll show you how to capture that ID shortly. 

Next, modify the <code>SubmitArtistLogic</code> struct in <code>endpoint/artist.go</code> so it looks like:

```go

```