# Tutorial - Service errors 

## What you'll learn

1. How to control the type of HTTP response your clients receive for different error conditions

## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Followed the [setting up a test database](006-database-read.md) section of [tutorial 6](006-database-read.md)
 1. Either have completed [tutorial 8](008-shared-validation.md) or open a terminal and run:
 
<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 9
</pre>


## Test database

If you didn't follow [tutorial 6](006-database-read.md), please work through the '[Setting up a test database](006-database-read.md)'
section which explains how to run Docker and MySQL with a pre-built test database.

## Error behaviour

The error conditions that a web service has to cope with can be broadly divided into four categories:

1. Security or access related errors, where a client tries to do something that they're not allowed to do.
1. Client errors, where the client has made a malformed web service call or supplied invalid data.
1. Logic errors, where the client's request is technically valid, but the the requested operation is not possible.
1. Unexpected errors, where the service has encountered an internal problem or a problem with a downstream dependency.

## Service errors

Granitic provides a mechanism for dealing with all these types of error in a consistent way - the [service error](https://godoc.org/github.com/graniticio/granitic/ws)

The [ws.WsResponse](https://godoc.org/github.com/graniticio/granitic/ws#WsResponse) object passed to your logic components' [process](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestProcessor)
method represents the data and state that will be returned to your client (via HTTP) when the request has been processed. It has a field:

```go
   Errors *ServiceErrors
```

that provides access to a [ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors) struct. Your code should
add errors to this structure as they are encountered. When processing is complete, Granitic will evaluate the errors in this structure to
determine what type of response and HTTP status code (200, 500 etc) should be sent to your client. 


To see this in action, edit your <code>ArtistLogic</code> struct in <code>endpoint/artist.go</code>
so the <code>Process</code> and change the two lines that look like:

```go
  res.HttpStatus = http.StatusInternalServerError
```

to

```go
  res.Errors.AddPredefinedError("DATABASE_UNEXPECTED")
```

Instead of telling Grantic to explicitly set the HTTP status to 500 (internal server error), we are instead logging that a particular error
has occured. The code we are passing to <code>AddPredefinedError</code> relates to an entry in your list of <code>serviceErrors</code> in your
<code>resource/config/config.json</code> file, which currently looks like:

```json
"serviceErrors": [
  ["C", "INVALID_ARTIST", "Cannot create an artist with the information provided."],
  ["C", "NAME_MISSING", "You must supply the Name field on your submission."],
  ["C", "NAME_BAD_LENGTH", "Names must be 5-50 characters in length."],
  ["C", "NAME_BAD_CONTENT", "Names can only contain letters and spaces."],
  ["C", "FIRST_ACTIVE_INVALID", "FirstYearActive must be in the range 1700-2100"],
  ["C", "NO_SUCH_RELATED", "Related artist does not exist"]
]
```

## Error categories

You'll notice that each entry has the string "C" before the error code. This indicates that the error is a 'client' error, caused by a mistake made
by the web service client. If Granitic finds only client errors in a response, the response's HTTP status will be set to <code>400 - Bad request</code>

We need to add another entry to that list:

```json
  ["U", "DATABASE_UNEXPECTED", "Unexpected problem communicating with the database"]
```

The type of error in this case is "U" or <code>unexpected</code>. If Granitic finds an 'unexpected' error in the response, it will set the HTTP status
code to <code>500 - Internal server error</code>. Determining which HTTP status code is set follows the rules defined here under [Http status code determination](https://godoc.org/github.com/graniticio/granitic/ws)

## Building and testing

Start your service:

```
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind && go build && ./recordstore -c resource/config
```

and stop your database (if you are using the example docker image run <code>docker stop graniticrs_recordstore-db_1</code> )

Sending the valid request:

```json
{
  "Name": "Another Artist"
}
```

will result in this response with an HTTP 500 status code


```json
{  
  "General":[  
    {  
      "Code":"U-DB_UNEXPECTED",
      "Message":"Unexpected problem communicating with the database"
    }
  ]
}
```




 
