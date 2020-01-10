# Enabling and defining validation rules
[Reference](README.md) | [Automatic Validation](vld-index.md)

## Enabling automatic validation
Automatic validation is configured on a [per-endpoint](ws-handlers.md) basis. There is no
specific [facility](fac-index.md) associated with validation, but each validation rule is
associated with an error code, which in turn is associated with an error message.

Management of error codes and messages is the responsibility of the [Service Error Management](fac-service-errors.md)
facility so this facility must be enabled.

### Configuring your handler

Your endpoint must be managed by an instance of [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler)
(as [described here](ws-handlers.md)) and be configured to provide a [data capture target](ws-capture.md) - this is the 
object that will actually be validated.

### Adding an Auto Validator

Automatic validation is enabled by injecting an instance of the built-in 
[validate.RuleValidator](https://godoc.org/github.com/graniticio/granitic/validate#RuleValidator)
struct into the `AutoValidator` field of your handler. This is generally performed inline like:

```json
"submitRecordHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "POST",
  "PathPattern": "^/record[/]?$",
  "Logic": {
    "type": "record.PostLogic"
  },
  "AutoValidator": {
    "type": "validate.RuleValidator",
    "DefaultErrorCode": "INVALID_RECORD",
    "Rules": "$createRecordRules"
  }
}
``` 

You must provide a `DefaultErrorCode` and a set of `Rules`. These are described below.

### Default error code

Every problem encountered during validation can be assigned an error code, which is mapped to a 
message by the [Service Error Management](fac-service-errors.md) facility. The default error
code is used when a more specific code is not set at the rule or operation level (see below).

## Rules

Validation rules are defined in your [configuration files](cfg-files.md) as a multi-dimensional
string array (`[][]string`) known as a rule set. They define which
fields on your [data capture target](ws-capture.md) you want to validate and a series of
[operations](vld-operations.md) to run against those fields.

An example rule set is:

```json
{
  "createRecordRules": [
    ["CatalogRef",  "STR",               "REQ:CATALOG_REF_MISSING", "HARDTRIM",        "BREAK",     "REG:^[A-Z]{3}-[\\d]{6}$:CATALOG_REF"],
    ["Name",        "STR:RECORD_NAME",   "REQ",                     "HARDTRIM",        "LEN:1-128"],
    ["Artist",      "STR:ARTIST_NAME",   "REQ",                     "HARDTRIM",        "LEN:1-64"],
    ["Tracks",      "SLICE:TRACK_COUNT", "LEN:1-100",               "ELEM:trackName"],
    ["Label",       "RULE:validLabel"]
  ]
}
```

Note that the formatting above is not required and just used here to improve readability.

## Rule structure

Rules consist of three components: 
 * The field name
 * A logical type
 * One or more operations

### Field name

The field name is the field in the [data capture target](ws-capture.md) that is to be validated.
The name must be an exact case sensitive match for the field on the target object.

### Field type

Field type is is not the Go type of the field to be validated, but a more generic concept of type that defines which 
[operations](vld-operations.md) are available.

Type must be one of the following:

| Type | Go type of field |
| ---- | ---- |
| STR | A `string` or a `*types.NilableString` |
| OBJ | Any `struct` or `map` |
| INT | An `int` of any size or signedness or a `*types.NilableInt64` |
| BOOL | A `bool` or a `*types.NilableBool` |
| FLOAT | A `float` of any size or signedness or a `*types.NilableFloat64` |
| SLICE | A slice or array of any type |

You may also set an error code after the type (e.g. `STR:INVALID_NAME`). This error code is
then used instead of the default error code defined for the rule set.

### Shared rules

Validation rules are generally specific to a specific [endpoint](ws-handlers.md), but some rules need to be shared
across multiple endpoints. The documentation for [shared rules](vld-custom.md) explains how this works.

### Operations

Operations are either checks that should be performed against the field (length checks, regular expressions etc), 
processing instructions (break processing if the previous check failed) or manipulations of the data to be validated 
(trim a string, etc). 

Many operations support arguments that are specified after the operation name. For example, `LEN:1-128` is an operation
on a string that requires the validated string to be between 1 and 128 characters in length. 

You may also set an error code after the operation name and/or arguments (e.g. `REQ:CATALOG_REF_MISSING` or 
`LEN:1-100:TRACK_LENGTH`). This error code is then used instead of the default error code defined for the rule or rule set.

The operations that are available for each supported type are documented in the [operations reference](vld-operations.md).

## Ordering

The ordering of rules and the operations within them is significant. Rules are applied to fields
in the same order in which they are specified, as are the operations within each rule.

---
**Next**: [Operations reference](vld-operations.md)

**Prev**: [Automatic validation principles](vld-principles.md)
