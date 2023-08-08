# Component definition files
[Reference](README.md) | [Component Container](ioc-index.md)

---

Component definition files are JSON files in which you declare which Go structs you want Granitic to manage as components,
the relationships between your components and how configuration should be found for your components. They are
loaded by the [grnc-bind](gpr-build.md) command line tool and transformed into Go source code. This process is known as
binding.


## File types, names and locations

Your component definition files must be valid UTF-8 encoded JSON files with a `.json` extension. You may split your 
components across as many files as you see fit.

You can store these files in any location that is accessible to the `grnc-bind` tool but component definition files
are generally considered part of of your application's build-time assets so would normally be checked into version control.

The default location for component definition files is in a folder called `comp-def` under your application's project folder.

## Merging

The [grnc-bind](gpr-build.md) command line tool loads all of your component definition files into memory and merges them
together. It is highly recommended that a component's declaration is contained within a single file. If you want to see
the effective merged component definition file that `grnc-bind` creates you can run [grnc-bind with the -m flag set](gpr-build.md)

## Packages

Like Go source files, your component definition files need to state which Go packages contain the types you want to 
use for your components. Packages are specified in a string array named `packages` at the root of your component definition
file's JSON structure.

```json
{
  "packages": [
    "github.com/graniticio/granitic/v2/ws/handler",
    "github.com/graniticio/granitic/v2/validate",
    "recordstore/artist",
    "recordstore/db",
    "github.com/go-sql-driver/mysql"
  ]
}
``` 

If you have multiple component definition files you may declare packages in each of them. The list of packages to 
import is merged together by `grnc-bind` and any duplicates are discarded.

## Aliases

If you import two or more packages where the final part of the package names clash (e.g `"myapp/handler"` and 
`github.com/graniticio/granitic/v2/ws/handler"`) you can assign an alias to one of the clashing packages in the JSON 
object/map `packageAliases` at the root of your component definition file's JSON structure:

```json
{
  "packageAliases": {
    "mh": "myapp/handler"
  }
}
```

Note that a package should be defined in _either_ the `packageAliases` section or the `packages` section, not both.

## Components

You define your application's components in the `components` JSON object/map at the root of your component definition file's JSON structure:

```json
{
  "components": {
    "myHandler": {
      "type": "my.Type"
    }
  }
}
```

The minimum requirements for a valid component are that it has a unique name (`myHandler` in the example above) and either 
a type (`my.Type` in the examples above) or a reference to a component template from which the type can be inferred (see below).

### Names

A component name must be unique within your application. Names must also follow the rules for Go variable name with one
exception - there is no significance attached to whether the first letter of the name is upper or lower case. Granitic
has a weak convention that component names start with a lower case letter, by you are free to ignore that convention.

### Types

Component types are formatted as the same way you would use an imported struct in a Go source file, with the last part of the containing package's name
following by the name of the type itself.

## Component configuration

After Granitic instantiates your component, it can inject values into any exported field on the underlying struct. 
Values can either be:

  1. Defined explicitly in a component's declaration
  1. Deferred with a 'configuration promise'
  
### Explicit values

Values that will not change according to the environment in which your application will run, and are not sensitive
in any way, can be directly specified as part of the component's declaration. 

For example:

```json
"submitArtistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "POST",
  "PathPattern": "^/artist[/]?$"
}
```

declares a [web service handler](ws-handlers.md) using Granitic's built-in `WsHandler` type.  This is an example of 
'static configuration' where your component's behaviour is configurable but will not change between environments 
(you will always want this handler to be an HTTP POST handler and match the same path pattern).

#### Security

It is important to note that configuration defined in component definition files will be compiled into your application
so it is strongly recommended that your do not specify passwords or other sensitive information in these files.


### Configuration promises

Configuration that will change between environments, is sensitive in some way or may need to be changed without
rebuilding your application should be defined in [files or URLs](cfg-files.md) that are made available to your
application _when it starts_. For each field that should be configured in this way you provide a _configuration promise_

For example:

```json
"dbConfig": {
  "type": "mysql.Config",
  "User": "$database.user",
  "Passwd": "$database.password",
  "Addr": "$database.host",
  "DBName": "$database.instance",
  "AllowNativePasswords": true
}    
```

declares a component that will store credentials and other configuration for accessing a database. Each field that
could change between environments is associated with a configuration promise. In this case we are promising that
a JSON configuration file will be available at application start time that has an object called `database` in
the root of the file with a series of fields containing actual configuration values. The route to a particular value (e.g
`database.instance` is called a _configuration path_)

Granitic will automatically inject the values found into your component.

#### The $ prefix

The `$` symbol in front of each configuration path tells `grnc-bind` that this string is a configuration promise and not
an explict string value that you want injected into your component. If you _do_ have a string value starting with a `$`,
you can escape it with `$$`

If you find the `$` notation difficult to read, you can use `conf:` or `c:` as an alternative.


#### Default values

Occasionally it is convenient to specify a value to be used if the configuration path cannot be found at application 
startup time. In these circumstances you can supply a default value in round brackets after the configuration promise.

For example:

```json
"dbConfig": {
  "Addr": "$database.host(localhost)"
}
```

This is not generally the recommend approach - instead you should maintain a [base configuration file](cfg-files.md) that
contains the default values for any configuration paths your application relies upon.


## References to other components

Granitic can link your components together by automatically injecting references. The only requirement is that the
_receiving_ component must have a field that is typed as a _pointer_ to the type that is being injected.

For example:

```go
type A struct {}

type B struct {
  ComponentA *A
}
```

### Explicit references

The most common way of linking two components is to explicitly declare the relationship in your component definition file.

For example:

```json
{
  "components": {
  
    "dbConfig": {
     "type": "mysql.Config"
    },
  
    "dbProvider": {
      "type": "db.MySQLProvider",
      "Config": "+dbConfig"
    }
  }
}
``` 

#### The + prefix

The `+` symbol tells Granitic that the contents of that string is the name of another component to be injected into
the field. In the above example, the instance of `mysql.Config` associated with the `dbConfig` component will be injected 
into the field `Config *mysql.Config` on the `dbProvider` component.

If you find the `+` symbol too subtle, you can use `ref:` or `r:` instead.


### Nested components

If you have a component that will only realistically be referenced by one other component, it is possible to 
_nest_ your component declarations to make this clear.

For example:

```json
"submitArtistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "POST",
  "PathPattern": "^/artist[/]?$",
  "Logic": {
    "type": "artist.PostLogic"
  }
}
```

In this example the component that is to be injected into the `submitArtistHandler`'s `Logic` field is being declared
inline.


#### Names allocated to nested components

Nested components are effectively anonymous so cannot be referred to by other components or manipulated through
[runtime control](rtc-index.md). You can explicitly assign a name to a nested component to get around this limitation.

For example:

```json
"submitArtistHandler": {
  "type": "handler.WsHandler",
  "Logic": {
    "type": "artist.PostLogic"
    "name": "submitLogic"
  }
}
```

The nested component in the above example now has the name `submitLogic`.


### Decorators

There are occasionally circumstances where references cannot or should not be explicitly defined in component definition
files. Granitic supports a pattern of decoration where references between components can be made programmatically. This
[decorator pattern is described here](ioc-decorators.md)


## Framework modifiers

In order to meet the principle of [easily replaced core features](int-index.md), Granitic provides a mechanism for 
modifying its internal components via configuration. This feature is known as *framework modification* and is 
considered an **advanced feature** and carries some risk as it makes your code dependent on internal Granitic naming and 
code structure that might be refactored in a later release.

Some sections of in this reference manual may suggest you use framework modification in lieu of more robust customisation
that is expected in a later release of Granitic.

### Modifying a component

Granitic manages its internal components in the same container as the user components that you create. By convention, these
components are named with the prefix `grnc`. In order to modify a framework component you must know its name. The [facility
documentation](fac-index.md) generally lists the name and type of the components created by each facility. 

Once you have the name of the component you want to modify, you need to add a new section to your component definition file 
(for this example we are modifying the `grncJSONResponseWriter` component that is part of the [JSONWs facility](fac-json-ws.md))

```json
"frameworkModifiers": {
  "grncJSONResponseWriter": {
    "StatusDeterminer": "myCustomHttpCodes"
  }
}
```

Here we are telling Granitic to override the component that is injected into the `StatusDeterminer` field of `grncJSONResponseWriter`
(which is an instance of [ws.MarshallingResponseWriter](https://godoc.org/github.com/graniticio/granitic/v2/ws#MarshallingResponseWriter))
with a different component. The component you inject can be any component that you have defined in the `components` section
of your component definition file.

Note that the name of the injected component is not prefixed with a `+`, which would normally indicate a reference to another
component. That is because framework modification _only_ supports the injection of components, not literal values or
configuration promises.


---
**Next**: [Component templates](ioc-templates.md)

**Prev**: [Component container principles](ioc-principles.md)

 