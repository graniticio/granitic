# Building and running a Granitic application
[Reference](README.md) | [Granitic Projects](gpr-index.md)

---

Building Granitic applications is a two stage process - bind then build. You can then run your application

## Binding

Granitic's core feature is a [component container](ioc-index.md) which allows Granitic to manage the lifecycle of 
those objects (structs) that a developer wants to treat as components. This is an implementation of the well-known 
[inversion of control and dependency injection patterns](https://martinfowler.com/articles/injection.html).

Refer to the [component container principles](ioc-principles) for more information about the concepts around components.

Developers define the behaviour and relationships between components in [JSON files](ioc-definition-files.md). 
In intepreted lanaguges (Python, Ruby etc) or virtual machine langauges (Java, .NET) these files could be loaded at 
runtime to discover which objects need to be created and, if those objects are not part of the base programming language, 
the libraries which _do_ contain the objects can be dynamically loaded.

This is not a realistic option in mainstream Go - some of Go's core benefits (small distributables, fast startup and 
execution times) derive from the fact that it is a generally used as a compiled and statically linked lanaguge.

This means that any code that you want to use in your application must be available when you run `go build`. To bridge 
the gap between the IoC container convention of defining components in configuration files and Go's static nature, Granitic 
requires an additional command `grnc-bind` to be run before `go build`. 

### Installing grnc-bind

`grnc-bind` is shipped with Granitic and can be built with:

```
  go install github.com/graniticio/granitic/v2/cmd/grnc-bind
```

As long as your `$GOPATH/bin` folder is in your `$PATH`, you will be able to run this tool from any folder

### What does grnc-bind do?

`grnc-bind` takes the definitions in your [component definition files](ioc-definition-files.md) and converts 
them into Go source code. By default the file containing the generated code is _bindings/bindings.go_

### When should I run grnc-bind?

`grnc-bind` must be run whenever you modify one of your [component definition files](ioc-definition-files.md). If you are using an IDE, it is recommend that your configure it to run `grnc-bind` automatically whenever a component definition file is saved.

It is also prudent to run it immediately before you run `go build`. A common pattern for building Granitic applications from the command line is:

```
  cd /path/to/project && grnc-bind && go build
```

### Command line options for grnc-bind

When you run `grnc-bind` without arguments, it assumes:

  1. You are the root folder of the project you want to build
  2. The project stores its component definition files under `resource/components/`
  3. The generated Go code will be stored under `bindings/bindings.go`

This behaviour can be modified by command line arguments:

```
Usage of grnc-bind:

  -c string
    	A comma separated list of component definition files or directories containing component definition files (default "comp-def")
  -l string
    	The level at which messages will be logged to the console (TRACE, DEBUG, WARN, INFO, ERROR, FATAL) (default "WARN")
  -m string
    	The path of a file where the merged component definition file should be written to. Execution will halt after writing.
  -o string
    	Path to the Go source file that will be generated to instatiate your components (default "bindings/bindings.go")
```

### Debugging grnc-bind

`grnc-bind` will exit with an error if your component definition files are syntactically or logically incorrect. However 
it is possible to define component definition files that are correct but cause invalid Go to be generated.

In this case, your IDE will indicate that the generated `bindings.go` file will not compile and the compilation 
errors for that file should show you what is wrong with your [component definition file](ioc-definition-files.md).

## Building and deploying

Once you have run `grnc-bind`, you can build your application like any other go application with `go build` or `go install`.

Apart from configuration files, this executable is the sole distributable
for your application. You can copy it to any machine running the same architecture as the one you built on and it will run. 
Go and Granitic _do not_ need to be installed on the host machine.

You can also use Go's cross-compilation support to compile your Granitic application for any target architecture.

## Running your application

Once you have built your application you will have an executable with the same name as the Go source file that contains 
your main method. If you have used the `grnc-project` tool to create your project, the executable will be 

`service` or `service.exe`

which can be run directly from the command line. If you run the executable without arguments, it will assume that 
you are running from the root project folder and that configuration for the application can be found in `config/`. 
This is normally the desired behaviour in development, but not in other environments. 
You can use built-in command line arguments to change this behaviour.

### Command line arguments for your application

Granitic applications automatically support a series of command line arguments to alter runtime behaviour:

#### Configuration file locations -c

The -c flag allows you to specify a comma separated list of configuration files, folders containing configuration files 
or URLs that will serve JSON files that will be merged together to form a single configuration for your application.

For example, if on your production machine your configuration files are stored in `/var/service-config` you would start 
your application with:

```
-c /var/service-config
```

Refer to the [config merging](cfg-merging.md) and [remote configuration](adm-remote.md) documentation for more details.


#### Startup framework log level -l

Granitic has a comprehensive [logging facility](log-index.md) for controlling the outout of both your application's 
logging and Granitic's built-in components. However this facility only starts _after_ your application's configuration 
files are parsed. 

If you need to debug what is happening inside Granitic before that point, you can use the the -l parameter with one 
of `TRACE, DEBUG, WARNING, ERROR, INFO, FATAL` to control framework logging during the 'bootstrap' phase.

#### Defer startup framework logging -d

Granitic logs some information before it has loaded your application's configuration and applied it to the [logging facility](log-index.md).
If you are using [custom log formatting](log-format.md), this means those early messages will not be formatted to your 
requirements. 

Providing the -d flag causes Granitic to hold all logging messages in memory until it has set up the logging facility.
 

#### Instance ID -i 

If you are hosting multiple instances of your Granitic application on a single server/VM, or just want to give a 
logical name to a particular instance, you can specify the name using the -i flag.

Refer to the [instance indentification](adm-instance.md) documentation for more details.



---
**Next**: [JSON and YAML](gpr-json.md) 

**Prev**: [Building and running your application](gpr-build.md)
