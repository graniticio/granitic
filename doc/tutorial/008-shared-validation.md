# Tutorial - Shared validation rules 

## What you'll learn

1. How to write validation rules that can be shared between multiple endpoints
1. How to delegate validation of a field to a Granitic component

## Prerequisites

 1. Follow the Granitic [installation instructions](../installation.md)
 1. Read the [before you start](000-before-you-start.md) tutorial
 1. Followed the [setting up a test database](006-database-read.md) section of [tutorial 6](006-database-read.md)
 1. Either have completed [tutorial 7](007-database-write.md)  or clone the
    [tutorial repo](https://github.com/graniticio/tutorial) and navigate to `json/008/recordstore` in your terminal.
 
## Test database

If you didn't follow [tutorial 6](006-database-read.md), please work through the '[Setting up a test database](006-database-read.md)'
section which explains how to run Docker and MySQL with a pre-built test database.

### Shared rules

The validation rules we've expressed in `config/base.json` currently look like:

```json
"submitArtistRules": [
  ["Name",             "STR",  "REQ:NAME_MISSING", "TRIM", "STOPALL", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
  ["FirstYearActive",  "INT",  "RANGE:1700|2100:FIRST_ACTIVE_INVALID"]
]
```

These rules are specific to submitting an artist, but some rules (like checking to see if an artist exists) are likely to 
be useful in a number of places. Granitic provides a mechanism for defining rules in a way in which they can be shared. Open 
`comp-def/common.json` and add this component:

```json
"sharedRuleManager": {
  "type": "validate.UnparsedRuleManager",
  "Rules": "$sharedRules"
}
```

In the same file, modify the `submitArtistValidator` component so its definition looks like:

```json
"submitArtistValidator": {
  "type": "validate.RuleValidator",
  "DefaultErrorCode": "INVALID_ARTIST",
  "Rules": "$submitArtistRules",
  "RuleManager": "+sharedRuleManager"
}
```

We now need to edit `config/base.json` to add some shared rules. Add the following:

```json
"sharedRules": {
  "artistExistsRule": ["INT", "EXT:artistExistsChecker"]
}
```

`EXT` (short for external) is an operation that delegates validation of a field to another Granitic component, in this case
a component named `artistExistsChecker` that will need to implement the 
[validate.ExternalInt64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalInt64Validator)
interface.

We need to alter the existing `submitArtistRules` in `config/base.json` so that they use the shared we've just defined when 
checking the set of 'related artists' that are provided when creating a new artist:

```json
"submitArtistRules": [
  ["Name",             "STR",  "REQ:NAME_MISSING", "TRIM", "STOPALL", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
  ["FirstYearActive",  "INT",  "RANGE:1700|2100:FIRST_ACTIVE_INVALID"],
  ["RelatedArtists", "SLICE",  "ELEM:artistExistsRule:NO_SUCH_RELATED"]
]
```

`ELEM` is an operation that causes a shared rule to be applied to each _element_ of a slice. We have introduced an new error code
`NO_SUCH_RELATED`, so we'll need to add that to our `serviceErrors` in `config/base.json`. Add the following:

```json
  ["C", "NO_SUCH_RELATED", "Related artist does not exist"]
```


### Optional exercise

You'll notice that the configuration file `config/base.json` and the component definition file `comp-def/common.json`
are getting quite complex now. Try refactoring these into multiple files (e.g `config/base.json`, `config/validation.json`,
`config/messages.json`)

### Validation component

We now need to build the component that actually performs the database check. Create a new file `db/validate.go` 
and set its contents to:

```go
package db

import (
  "github.com/graniticio/granitic/v3/rdbms"
  "github.com/graniticio/granitic/v3/logging"
)

type ArtistExistsChecker struct{
  DbClientManager rdbms.ClientManager
  Log logging.Logger
}

func (aec *ArtistExistsChecker) ValidInt64(id int64) (bool, error) {

  dbc, _ := aec.DbClientManager.Client()

  var count int64

  // Execute a query that counts how many artists in the database share the ID we are checking
  // If the count is zero, that artist doesn't exist.
  if _, err := dbc.SelectBindSingleQIDParam("CHECK_ARTIST", "ID", id, &count); err != nil {
    return false, err
  } else {
    return count > 0, nil
  }
}
```

An we'll need to add a new query to `resource/queries/artist`:

```sql
ID:CHECK_ARTIST

SELECT
    COUNT(id)
FROM
    artist
WHERE
    id = ${ID}
```

The last step is to register this new checker as a component by adding the following to your `comp-def/common.json` file:

```json
"artistExistsChecker": {
  "type": "db.ArtistExistsChecker"
}
```

### Binding database results to a basic type

Previous examples have shown how to bind database results into a struct or slice of structs. In the Go code above

```go
  dbc.SelectBindSingleQIDParam("CHECK_ARTIST", "ID", id, &count)
```

we are binding the results of the database call to an int64. You may supply a basic type (string, int etc) instead of a
struct when your query is _guaranteed_ to return a single row with a single column.

## Building and testing

Start your service by navigating to the folder where you have yout tutorial project and running:

```
grnc-bind && go build && ./recordstore
```

and POST the following JSON to `http://localhost:8080/artist`

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
 