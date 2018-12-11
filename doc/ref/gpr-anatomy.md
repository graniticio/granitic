# Anatomy of a Granitic application

Granitic applications are built from three main resources:

  * Go source code
  * Component definition files
  * Configuration
  
## Go source code

Your application will define its application logic in Go source files. There is no restriction
on how you structure your code in terms of packages and source files. The only two restrictions are:

 1. Your application needs a main method in which it can pass control to Granitic
 2. The Granitic IoC container expects to manage components that are instances of [structs](https://gobyexample.com/structs)
 so your application logic needs to be represent as structs.
 
 ## Component definition files
 
 A component is a named instance of a struct that Granitic will instantiate and configure on behalf of
 your application. Component definition files are JSON ([or YAML](https://github.com/graniticio/granitic-yaml)) files
 that contain the name, type and some of the configuration for that component. For example:
 
 ```json
  "artistHandler": {
      "type": "handler.WsHandler",
      "HttpMethod": "GET",
      "Logic": "ref:artistLogic",
      "PathPattern": "^/artist"
    }
```

is the definition for a _handler_ component that can receive web service requests.

The structure and capabilities of these files are described [in detail here](020-component-definition-files.md)

### Binding

Building Granitic applications requires a step called _binding_ this is the process of converting the component
definitions into Go source files using the built-in [grnc-bind](apx-001-command-line-tools.md) command. For most projects 
the build process is simply:

```
    grnc-bind && go build
```