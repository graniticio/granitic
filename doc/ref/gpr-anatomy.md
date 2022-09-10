# Anatomy of a Granitic application
[Reference](README.md) | [Granitic Projects](gpr-index.md)

---

Granitic applications are built from three main resources:

  * Go source code
  * Component definition files
  * Configuration files
  
## Go source code

Your application will define its application logic in Go source files. There is no restriction
on how you structure your code in terms of packages and source files. The only two requirements are:

 1. Your application needs a main method in which it can pass control to Granitic
 2. The Granitic IoC container expects to manage components that are instances of [structs](https://gobyexample.com/structs) 
 so your application logic needs to be represented as structs with member functions.
 
### Minimal 'main' file

```go
package main

import "github.com/graniticio/granitic/v3"
import "project/bindings"  

func main() {
	granitic.StartGranitic(bindings.Components())
}
```

This example file provides a `main` function for Go and passes control to Granitic using the `StartGranitic` function. 
This allows Granitic to have full control over the parsing of [command line arguments](gpr-build.md). 
If your application needs to parse it's own command line arguments, refer to the GoDoc for 
[alternative ways to start Granitic](https://godoc.org/github.com/graniticio/granitic)

## Component definition files
 
A component is an instance of a struct that Granitic will instantiate and configure on behalf of
your application. Component definition files are JSON ([or YAML](https://github.com/graniticio/granitic-yaml)) files
that contain the name, type and some of the configuration for that component. 

For example:
 
 ```json
  "artistHandler": {
      "type": "handler.WsHandler",
      "HTTPMethod": "GET",
      "Logic": "ref:artistLogic",
      "PathPattern": "^/artist"
    }
```

is the definition for a _handler_ component that can receive web service requests. Your application _must_ include at 
least one component definition file.

The structure and capabilities of these files are described [in detail here](ioc-definition-files.md)

## Configuration files

Granitic applications are designed such that minimal configuration is embedded in the compiled executable created when you 
`go build` your application. Instead you provide configuration as JSON files that are loaded when your application starts. 
The usage of [configuration files is described in detail here](cfg-index.md).

## Default locations

By default, Granitic expects to find the following directory structure in your project root:

```
  comp-def
    files.json
  config
    files.json
```
Where component definition files are stored in the `comp-def` directory and configuration files are stored in `config`

But this is only a convention and the tools that interact with component definition files and configuration files 
accept alternative locations for these files.

---
**Next**: [Creating a new project](gpr-create.md)

**Prev**: [Granitic projects](gpr-index.md)




