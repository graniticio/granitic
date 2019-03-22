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

You define components in the `components` JSON object/map at the root of your component definition file's JSON structure:

```json
{
  "components": {
    "myHandler": {
      "type": "my.Type"
    }
  }
}
```

The minimum requirements for a valid component is that it has a unique name (`myHandler` in this case) and either 
a type or a reference to a component template from which the type can be infered (see below)

### Names

A component name must be unique within your application. Names must also follow the rules for Go variable name with one
exception - there is no significance attached to whether the first letter of the name is upper or lower case. Granitic
has a weak convention that component names start with a lower case letter, by you are free to ignore that convention.

### Types

You specify a type in the same way you would in a Go source file - with the last part of the containing package's name
following by the name of the type itself.

The type must be an exported struct.
