# Tutorial - Shared validation rules 

## What you'll learn

1. How to write validation rules that can be shared between multiple endpoints
1. How to delegate validation of a field to a Granitic component

## Prerequisites

 1. Follow the Granitic [installation instructions](../installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Followed the [setting up a test database](006-database-read.md) section of [tutorial 6](006-database-read.md)
 1. Either have completed [tutorial 7](007-database-write.md) or open a terminal and run:
 
<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 8
</pre>


## Test database

If you didn't follow [tutorial 6](006-database-read.md), please work through the '[Setting up a test database](006-database-read.md)'
section which explains how to run Docker and MySQL with a pre-built test database.

### Shared rules

The validation rules we've expressed in <code>resource/config/config.json</code> currently look like:

```json
"submitArtistRules": [
  ["Name",             "STR",  "REQ:NAME_MISSING", "TRIM", "STOPALL", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
  ["FirstYearActive",  "INT",  "RANGE:1700|2100:FIRST_ACTIVE_INVALID"]
]
```

These rules are specific to submitting an artist, but some rules (like checking to see if an artist exists) are likely to 
be useful in a number of places. Granitic provides a mechanism for defining rules in a way in which they can be shared. Open 
<code>resource/components/components.json</code> and add this component:

```json
"sharedRuleManager": {
  "type": "validate.UnparsedRuleManager",
  "Rules": "conf:sharedRules"
}
```

In the same file, modify the <code>submitArtistValidator</code> component so its definition looks like:

```json
"submitArtistValidator": {
  "type": "validate.RuleValidator",
  "DefaultErrorCode": "INVALID_ARTIST",
  "Rules": "conf:submitArtistRules",
  "RuleManager": "ref:sharedRuleManager"
}
```

We now need to edit <code>resource/config/config.json</code> to add some shared rules. Add the following:

```json
"sharedRules": {
  "artistExistsRule": ["INT", "EXT:artistExistsChecker"]
}
```

<code>EXT</code> (short for external) is an operation that delegates validation of a field to another Granitic component, in this case
a component named <code>artistExistsChecker</code> that will need to implement the [validate.ExternalInt64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalInt64Validator)
interface.

We need to alter the existing <code>submitArtistRules</code> in <code>config.json</code> so that they use the shared rule on ID we're given:

```json
"submitArtistRules": [
  ["Name",             "STR",  "REQ:NAME_MISSING", "TRIM", "STOPALL", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
  ["FirstYearActive",  "INT",  "RANGE:1700|2100:FIRST_ACTIVE_INVALID"],
  ["RelatedArtists", "SLICE",  "ELEM:artistExistsRule:NO_SUCH_RELATED"]
]
```

The <code>ELEM</code> is an operation that causes a shared rule to be applied to each element of a slice. We have introduced an new error code
<code>NO_SUCH_RELATED</code>, so we'll need to add that to our <code>serviceErrors</code> in <code>config.json</code>. Add the following:

```json
  ["C", "NO_SUCH_RELATED", "Related artist does not exist"]
```

### Validation component

We now need to build the component that actually performs the database check. Create a new file <code>recordstore/db/validate.go</code> and set its contents to:

```go
package db

import (
  "github.com/graniticio/granitic/v2/rdbms"
  "github.com/graniticio/granitic/v2/logging"
)

type ArtistExistsChecker struct{
  DbClientManager rdbms.RdbmsClientManager
  Log logging.Logger
}

func (aec *ArtistExistsChecker) ValidInt64(id int64) (bool, error) {

  dbc, _ := aec.DbClientManager.Client()

  var count int64

  if _, err := dbc.SelectBindSingleQIDParam("CHECK_ARTIST", "ID", id, &count); err != nil {
    return false, err
  } else {
    return count > 0, nil
  }
}
```

An we'll need to add a new query to <code>resource/queries/artist</code>:

```sql
ID:CHECK_ARTIST

SELECT
    COUNT(id)
FROM
    artist
WHERE
    id = ${ID}
```

The last step is to add the following to your <code>components.json</code> file:

```json
"artistExistsChecker": {
  "type": "db.ArtistExistsChecker"
}
```

### Binding database results to a basic type

Previous examples have shown how to bind database results into a struct or slice of structs. In the above

```go
  dbc.SelectBindSingleQIDParam("CHECK_ARTIST", "ID", id, &count)
```

we are binding the results of the database call to an int64. You may supply a basic type (string, int etc) instead of a
struct when your query is guaranteed to return a single row with a single column.

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
  "RelatedArtists": [-1, 1, 9999]
}
```

(see the [data capture tutorial](004-data-capture.md) for instructions on using a browser plugin to do this) and you should
get a result like:

```json
{
  "ByField":{
    "RelatedArtists[0]":[
      {
        "Code":"C-NO_SUCH_RELATED",
        "Message":"Related artist does not exist."
      }
    ],
    "RelatedArtists[2]":[
      {
        "Code":"C-NO_SUCH_RELATED",
        "Message":"Related artist does not exist."
      }
    ]
  }
}
```


## Recap

 * Validation rules can be defined globally so they can be re-used by multiple endpoints
 * When validating slices, the validation of each element can be delegated to another validation rule
 * When validating ints, floats and strings your validation rule can delegate to another component, as long as it implements
 [validate.ExternalInt64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalInt64Validator), [validate.ExternalFloat64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalFloat64Validator)
  or [validate.ExternalStringValidator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalStringValidator)
 * Database results can be bound to a basic type as long as your query returns one row with one column.
 