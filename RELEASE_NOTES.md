# Granitic 2.0 Release Notes

Granitic 2.0 is a major release of Granitic focusing on YAML support, Go module support, streamlining of web service code, improvements
to `grnc-bind`, code quality and documentation.

This release is not backwards compatible, refer to the migration guide at the bottom of these notes for more information.


## 2.0.x fixes

See the section at the end of this document for fixes made since the initial 2.0.0 release

## YAML

Granitic now supports YAML for configuration and component definition files. This involves the
use of a third-party YAML parser and so to preserve Granitic's principle of having no
dependencies, you must download the additional [granitic-yaml](https://github.com/graniticio/granitic-yaml)
project from GitHub or add it as a Go module dependency (see below).

The default file format for Granitic is currently still JSON and all documentation an examples
will continue to use JSON, but YAML versions of the tutorial source code [are now available](https://github.com/graniticio/tutorial).

Note you will need to change your application's main function to 

```go
func main() {
	granitic-yaml.StartGraniticWithYaml(bindings.Components())
}
```

### New default locations for files

The previous default locations for application configuration (`resource/config`) and component definitions (`resource/components`)
were too verbose and have been deprecated. The new preferred locations are `config` and `comp-def`. 

`grnc-project` has been modified to create projects with these new locations. The empty files it creates are now `config/base.json` and
`comp-def/common.json` and the generated entrypoint file is `main.go`

The old locations are still respected, but you will see a warning when you use them. Support for the old locations will 
be removed in a future version of Granitic.


## Reference documentation

Granitic now has a reference manual, intended to compliment the information in the [Godoc](https://godoc.org/github.com/graniticio/granitic). 
You can find this manual in  `doc/ref/index.md` or on the [Granitic website](http://www.granitic.io/ref/).

## Go modules

Grantic 2 is compatible with the requirements of [Go modules](https://github.com/golang/go/wiki/Modules) including 
[semantic import verisioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning). As such Granitic
now requires the use of Go 1.11 or later.

Your applications should declare their dependency on Granitic in their `go.mod` file with:

```
    require github.com/graniticio/granitic/v2 v2
```

or (if you are using [YAML configuration and component](https://github.com/graniticio/granitic-yaml) files)

```
    require github.com/graniticio/granitic-yaml/v2 v2
```

Wherever your code (or component definition files) imports Granitic types, the import statements should be of
the form:

```go
    import "github.com/graniticio/granitic/v2/packageName"
```


### Tool support

`grnc-project` and `grnc-yaml-project` now generate working Granitic projects using modules. A `go.mod` will automatically
be created for you.

`grnc-bind` and `grnc-yaml-bind` have additional support for modules. These features are explained below.

## Web service streamlining

Your web service logic component no longer needs to implement [handler.WsRequestProcessor](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestProcessor) or 
[handler.WsUnmarshallTarget](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsUnmarshallTarget). Instead it just needs
to declare a method with the signature:

```go
  ProcessPayload(context.Context, *ws.Request, *ws.Response, *YourStruct)  
```

Where `*YourStruct` is a pointer to any type that you want the HTTP request's payload to be parsed into. 

If you do not need any body, path or query parameter data parsed (e.g. a simple GET or HEAD), you should continue to implement
[handler.WsRequestProcessor](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestProcessor). 

If you need more control over the struct that is created as your parsing target (e.g. you need to pre-populate it), you
should continue to implement [handler.WsUnmarshallTarget](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsUnmarshallTarget).

## Request instrumentation

Granitic now includes integration points for adding instrumentation to a web service request and
supporting for propagating that instrumentation to downstream web service requests. A common
use-case for this is to add timing traces to service calls.

See the [instrument](https://godoc.org/github.com/graniticio/granitic/instrument) package documentation for more details.

## Request unique identifiers

Granitic now includes integration points for generating unique IDs for web service requests
and having them injected into that request's `context.Context`. You need to create a component that implements
[IdentifiedRequestContextBuilder](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#IdentifiedRequestContextBuilder)

## Code quality

Granitic 2 sees significant improvements in file and statement unit test coverage. This release
also see the start of Granitic committing to abiding by _all_ of the advice from the `go vet` and 
`golint` tools.

This has resulted in changes to the names of a number of exported types and the names of some Granitic
facilities. See the migration guide at the bottom of these notes for information on how this might affect your
application.

## RDBMS client is now an interface

To make it easier to provide mock implementations for testing, the struct rdbms.RdbmsClient is now the interface
rdbms.Client

## grnc-bind

`grnc-bind` (and its YAML equivalent `grnc-yaml-bind`) have gained a number of additional features
and quality of life improvements.

### Validation and error handling

Rather than exit as soon as an error is found,`grnc-bind` attempts to continue parsing your
component definition files for as long as possible. This means that multiple errors are now reported
in a single run.

Additional validation has been applied to increase the number of problems detected during the bind
phase instead of during `go build`

### Logging

You can now provide a `-l LOGLEVEL` flag to `grnc-bind` for more detailed output. Valid
values for `LOGLEVEL` are `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` and `FATAL` (case
insensitive). Default is `WARN`

### New symbols for dependency and configuration promises

In addition to the `c:`, `conf:`, `r:` and `ref:` prefixes for indicating configuration promises
and component dependencies, you can now also use the symbols `$` and `+` as an alternative.

For example:

```json
"ComponentName": {
  "type": "some.Type",
  "A": "$some.config.path",
  "B": "+someOtherComponent"
}
```

If strings in your component definition file start with these new symbols, you can escape them with `$$` or `++`

### Nested components

You can now define components in a nested manner under the field of the parent component
that the nested component should be injected into.

For example:

```json
  "artistHandler": {
    "type": "handler.WsHandler",
    "PathPattern": "$paths.getArtist(^/artist)",
    "HTTPMethod": "GET",
    "Logic": {
      "type": "endpoint.ArtistLogic"
    }
}
```

The only requirement is that you specify a type or a parent template for the nested component.
If you want to be able to refer to the nested component from other components, you need to
provide a `name` field, e.g.:

```json
   "Logic": {
     "type": "endpoint.ArtistLogic",
     "name": "artistLogic"
   }
       
```


### Default values

You can now provide a default value along with a configuration promise by including the 
default in brackets after the config path. This value will be used if no configuration 
is provided to your application that overrides that config path.

For example:

```json
"ComponentName": {
  "type": "some.Type",
  "A": "$some.config.path(true)",
  "B": "$some.other.path(1.2"
}
```

Note that default values will only be type checked against the fields you are trying to
inject them into at application run time (not during the build phase).


### Package aliases

If your component definition files import two packages that clash (because the final part of the
package name is the same), you can define an alias for one of the packages, then refer to the
alias when defining types.

For example:

```json
{
  "packages": [
    "github.com/graniticio/granitic/v2/ws/handler"
  ],

  "packageAliases": {
    "mh": "myproject/handler"
  },
  
  "SomeComponent": {
    "type": "mh.SomeType"  
  }
}
```

### Find Granitic installation from go.mod (experimental)

`grnc-bind` serialises a copy of Granitic's facility configuration files into your application.
As such it needs to know where to find a copy of Granitic on disk. Previously this required
you to set the `$GRANITIC_HOME` environment variable or check Granitic out to a standard location.

In Granitic 2, the go.mod file can instead be used to automatically work out where to find
Granitic (but you still need to set your `$GOPATH` environment variable correctly).

This feature has not been fully tested with module proxies so should be considered experimental.

## Bug fixes

 * Component definition files containing a slice of `float64` could not be bound if the first element was parseable as an int.
 * The `ServiceErrorManager` was not respecting the value of `ErrorCodeUser.ValidateMissing()` when complaining about missing error codes.
 * Using `ConfigAccessor` to try and push configuration into an unsupported type of target field was not returning an error.
 * Some configuration parsing errors were causing Granitic to exit rather than return an error

# Granitic 1.x to 2.0 migration

Support for [Go modules](https://github.com/golang/go/wiki/Modules) and implementing the recommendations
of `golint` means that you will need to make changes to your application in order to run it with
Granitic 2.

## Import paths

[Semantic import verisioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning) support
means that you need to include the _major_ version number of Granitic in your import paths.
This will affect your Go source co and your component definition files.

### Go code

```go
  import "github.com/graniticio/granitic/logging"
```

would need to change to 

```go
  import "github.com/graniticio/granitic/v2/logging"
```

### Component definition files

```json
{
  "packages": [
    "github.com/graniticio/granitic/ws/handler"
  ]
}
```

would need to change to:

```json
{
  "packages": [
    "github.com/graniticio/granitic/v2/ws/handler"
  ]
}
``` 

## Capitalisation of abbreviations

Common abbreviations in type names (`Id`, `Http`, `Sql`, `Json`, `Xml`) are now capitalised
(`ID`, `HTTP`, `SQL`, `JSON`, `XML`)

For example the type `xml.GraniticXmlErrorFormatter` is now `xml.GraniticXMLErrorFormatter`

### Facilities

This has also affected the names of the facilities `HttpServer`, `JsonWs` and `XmlWs` which
are now `HTTPServer`, `JSONWs` and `XMLWs` respectively. You will need to update any references to 
these in your application's _configuration_ files.


## Stuttering

Types where the type name started with the name of the package have been 'de-stuttered'. For example
`ws.WsRequest` is now `ws.Request`.

This is most likely to affect your application code where it uses Granitic types in the `ws` and `rdbms`
packages.

## Default file locations

Move your configuration files from `resource/config` to `config` and your component definition files from `resource/components`
to `comp-def`

# Post release fixes

## 2.0.2

  * [rdbms.Client missing InsertQIDParam method](https://github.com/graniticio/granitic/issues/35)

