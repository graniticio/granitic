# Tutorial - Validating data submitted by a client

## What you'll learn

1. How to define rules to automatically validate incoming data
1. How to customise the error messages that are returned to your web service clients

## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 4](004-data-capture.md) or open a terminal and run

<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 5
</pre>

## Validation

Making sure that data submitted to a web service is safe and correct before processing it is a vital step. Even if your 
web services are for internal use only, you should never trust callers to always provide valid data. 

Performing validation in code can be a tedious process, generating lots of boilerplate code that can be difficult to read.
Granitic tries to make your code cleaner and your validation logic easier to understand by defining validation as rules
in your configuration files.

## Defining rules

We will be adding validation to the <code>/artist POST</code> endpoint we built in the [previous tutorial](004-data-capture.md)

Open your <code>/resource/config/config.json</code> file and set its contents to:

```json
{
  "Facilities": {
    "HttpServer": true,
    "JsonWs": true,
    "RuntimeCtl": true,
    "ServiceErrorManager": true
  },

  "ApplicationLogger": {
    "GlobalLogLevel": "INFO"
  },

  "environment": {
    "label": "DEV"
  },

  "submitArtistRules": [
    ["Name",  "STR",  "REQ"]
  ]
}
```
Rules are expressed as a JSON array of string arrays. Each entry in the top level array maps to the name of a field that we want to 
validate on our web service endpoint's [unmarshalling target](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsUnmarshallTarget) in this
case our target is the <code>SubmittedArtistRequest</code> which currently looks like:

```go
type SubmittedArtistRequest struct {
  Name *types.NilableString
  FirstYearActive  *types.NilableInt64
}
```

We have defined a single rule:

```json
  ["Name",  "STR",  "REQ"]
```

The first element in the array is the field we're checking on the target object. The second is the expected data type. The rest of the elements are 
the validation _operations_ that should be applied to the field. 

In this cause we are stating that a single operation should be applied. <code>REQ</code> means the field is required and validation will fail if it is _not provided in the JSON body_ (this is different to the field being present but empty)

We will cover many of the available operations throughout this tutorial, but if you'd like to skip ahead, the [validation GoDoc](https://godoc.org/github.com/graniticio/granitic/validate)
provides a complete reference.


### Service Error Manager

You might have noticed that we've enabled a new Granitic facility:

```json
   "ServiceErrorManager": true
```

this facility manages error messages for web service calls. We'll start using this shortly, but if you'd like to find out more, [refer to the GoDoc](https://godoc.org/github.com/graniticio/granitic/facility/serviceerror)

## Enabling validation

Validation is triggered by the handler component associated with your endpoint. In this case, it's the <code>submitArtistHandler</code> defined in your
<code>resources/components/components.json</code> file. We'll need to modify this component's configuration and also declare a new component:

```json
"submitArtistHandler": {
  "type": "handler.WsHandler",
  "HttpMethod": "POST",
  "Logic": "ref:submitArtistLogic",
  "PathPattern": "^/artist[/]?$",
  "AutoValidator": "ref:submitArtistValidator"
},

"submitArtistValidator": {
  "type": "validate.RuleValidator",
  "DefaultErrorCode": "INVALID_ARTIST",
  "Rules": "conf:submitArtistRules"
}
```

We've created a new component <code>submitArtistValidator</code> of type [validate.RuleValidator](https://godoc.org/github.com/graniticio/granitic/validate#RuleValidator). 
This is the code that will actually apply the validation rules. We've given it a reference to the validation rules we've defined in our config:

<pre>"Rules": "conf:submitArtistRules"</pre>

The <code>DefaultErrorCode</code> is used to lookup an error message to return if there isn't a specific error message defined for a particular validation operation.

[validate.RuleValidator](https://godoc.org/github.com/graniticio/granitic/validate#RuleValidator) is in a package that we haven't referenced before in our <code>components.json</code> file,
so you'll have to amend the <code>packages</code> array at the top of the file to include:

<pre>"github.com/graniticio/granitic/validate"</pre>

## Setting a default error message

Earlier we enabled the [Service Error Manager](https://godoc.org/github.com/graniticio/granitic/facility/serviceerror) facility. This facility
will refuse to start if we haven't defined error messages for each error code in use by our code. We've defined the code <code>INVALID_ARTIST</code>, so
open your <code>config.json</code> and add the following:

```json
"serviceErrors": [
  ["C", "INVALID_ARTIST", "Cannot create an artist with the information provided."]
]
```

The "C" indicates that this type of error is a client error. The types of error found during a web service call affect the
HTTP status code returned by the web service call. More information can be found [in the GoDoc](https://godoc.org/github.com/graniticio/granitic/ws)

### Service Errors config path

<code>serviceErrors</code> is the config path where the [Service Error Manager](https://godoc.org/github.com/graniticio/granitic/facility/serviceerror) facility expects
to find your error message defintions. If you want to use a different config path, you need to override the config value <code>ServiceErrorManager.ErrorDefinitions</code>

## Testing the first validation rule

If you haven't already done so, please look at the _Testing POST services_ section in the [previous tutorial](004-data-capture.md) which
explains how to use a browser plugin to test web service POST endpoints.

Start your service:

<pre>
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind && go build && ./recordstore -c resource/config
</pre>

And POST the following JSON to <code>http://localhost:8080/artist</code>

```json
{
  "FirstYearActive": 2010
}
```

You should see a response similar to:

```json
{
  "ByField":{
    "Name":[
     {
       "Code":"C-INVALID_ARTIST",
       "Message":"Cannot create an artist with the information provided."
     }
    ]
  }
}
```

and the HTTP status code will be <code>400 Bad Request</code> You can see that Granitic shows you the names of the fields that errors were
found on and a list of each error found on the field. If this is too verbose for your application, you can control how
error responses are formatted by providing your own [ws.ErrorFormatter](https://godoc.org/github.com/graniticio/granitic/ws#ErrorFormatter)


### Debugging validation

If you set <code>ApplicationLogger.GlobalLogLevel</code> or your <code>submitArtistValidator</code> component's
log level to <code>DEBUG</code> the console will log each operation run on a field. Refer to the [logging tutorial](003-logging.md)
to learn how to do this.

## Refining validation

At the moment, you could submit a zero-length <code>Name</code> and the service would accept it, which is 
probably not what we want. Change your validation rule to look like:

```json
  ["Name",  "STR",  "REQ", "LEN:5-50"]
```

This will reject the POST if the artist's name is fewer than five characters or greater than 50. The following JSON will be accepted:

```json
{
  "Name": "Some Artist",
  "FirstYearActive": 2010
}
```

### Whitespace around strings

The above rule would accept a name made up of five space characters. Try submitting:
```json
{
  "Name": "     ",
  "FirstYearActive": 2010
}
```

Granitic validation provides an operation <code>TRIM</code> which means that a string will have leading and trailing
whitespace removed before any other checks take place. Change your rule to:

```json
  ["Name",  "STR",  "REQ", "TRIM", "LEN:5-50"]
```

 and try again (the <code>Name</code> should be rejected).
 
 This <code>TRIM</code> operation does not permanently alter the data that is submitted to your service. If you submit a <code>Name</code> of "<code> ABCD </code>" 
 it will be validated as "<code>ABCD</code>" but will be passed to the rest of your code as "<code> ABCD </code>". String validation supports
 an alternative operation <code>HARDTRIM</code> that permanently trims the supplied data.
 
## Error codes and messages
 
At the moment our code just returns the same generic error code and message regardless of the problem. You can configure
Granitic to return a specific code and message for each problem.

Change your rule to:
 
 ```json
   ["Name",  "STR",  "REQ:NAME_MISSING", "TRIM", "LEN:5-50:NAME_BAD_LENGTH"]
 ```
 
 and your <code>serviceErrors</code> to
 
 ```json
 "serviceErrors": [
   ["C", "INVALID_ARTIST", "Cannot create an artist with the information provided."],
   ["C", "NAME_MISSING", "You must supply the Name field on your submission."],
   ["C", "NAME_BAD_LENGTH", "Names must be 5-50 characters in length."]
 ]
 ```
 
 Now each check on <code>Name</code> has a specific error message, as well as a catch-all message for any other errors found on the submitted data.
 
## More detailed rules and halting validation
 
Fields are checked in the order they are specified in your JSON configuration. Let's add a second rule that checks that
the optionally supplied <code>FirstYearActive</code> is in a reasonable range and change the first rule so that only letters and spaces are allowed in names:

 ```json
    ["Name",  "STR",  "REQ:NAME_MISSING", "TRIM", "LEN:5-50:NAME_BAD_LENGTH", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
    ["FirstYearActive",   "INT",  "RANGE:1700|2100:FIRST_ACTIVE_INVALID"]
 ```
 
We've added a new check on <code>Name</code> to make sure it matches the regular expression <code>^[A-Z]| +$</code> (letters or spaces only).
We've added a second rule to make sure that <code>FirstYearActive</code> is in the range 1700 to 2100 (note the pipe separator). We've also added two new error codes, so add the following
messages to your <code>serviceErrors</code>:

 ```json
  ["C", "NAME_BAD_CONTENT", "Names can only contain letters and spaces."],
  ["C", "FIRST_ACTIVE_INVALID", "FirstYearActive must be in the range 1700-2100"]
```
 
If you restart your service and send the following JSON:
 
```json
{
  "Name": "",
  "FirstYearActive": -1
}
```

You'll see that <code>Name</code> fails two tests - the length check and the regex check. It doesn't make sense to execute the <code>REG</code> operation
if the <code>LEN</code> operation has already failed. Change the rule to:

 ```json
    ["Name",  "STR",  "REQ:NAME_MISSING", "TRIM", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"],
    ["FirstYearActive",   "INT",  "RANGE:0|150:FIRST_ACTIVE_INVALID"]
 ```
 
 and you'll see that only the <code>LEN</code> operation applies. <code>BREAK</code> stops executing the current rule if the previous operation failed
 and moves on to the next rule.
 
 Sometimes a check is so fundamental that there is no point processing any further rules. Change the <code>Name</code> rule to:
 
 ```json
    ["Name",  "STR",  "REQ:NAME_MISSING", "TRIM", "STOPALL", "LEN:5-50:NAME_BAD_LENGTH", "BREAK", "REG:^[A-Z]| +$:NAME_BAD_CONTENT"]
```
 
and notice that the <code>FirstYearActive</code> checks are no longer executed if any problems are found with <code>Name</code>

## Advanced validation

This tutorial covers some of the most commons use cases for automatic validation, but Granitic offers much greater depth including
invoking components to validate fields, checking for mutual exclusivity between fields and defining callbacks to be invoked at certain points
during validation. Refer to the _Further reading_ section below for links to the GoDoc where these features are explained.

## Recap

 * Granitic can validate received data using rules defined in JSON configuration.
 * Error messages can be as specific as you choose - a message for each check, for each field or just one generic message.
 * Granitic's default JSON structure for error responses can be overridden.
 * Each validation rule is associated with a field on the web service's 'unmarshalling target'.
 * Rules are built from a series of validation operations.
 * Individual rules or the entire set of rules can be configured to halt when an error is found.
 
## Further reading

 * [Validation GoDoc](https://godoc.org/github.com/graniticio/granitic/validate)
 * [WsPreValidateManipulator](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsPreValidateManipulator) optional callback invoked before auto validation
 * [WsRequestValidator](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestValidator) optional callback invoked after auto validation has passed

 
 
## Next

The next tutorial covers the [reading of data from an RDBMS](006-database-read.md)