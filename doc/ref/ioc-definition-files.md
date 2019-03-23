# Component definition files

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
