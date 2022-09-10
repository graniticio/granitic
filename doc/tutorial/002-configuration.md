# Tutorial - Configuration

## What you'll learn

 1. How Granitic defines configuration
 2. How to define different configuration files for different environments
 3. How JSON configuration files are merged together

 
## Prerequisites

 1. Follow the Granitic [installation instructions](../installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 1](001-fundamentals.md) or clone the [tutorial repo](https://github.com/graniticio/tutorial)
 and navigate to `json/002/recordstore` in your terminal.

## Related GoDoc

https://godoc.org/github.com/graniticio/granitic/config

## Configuration

Granitic applications use JSON files to store configuration which is loaded when the application starts. Any valid JSON 
file is a valid configuration file:


```javascript
{
    "exampleBool": true,
    "exampleNumber": 1.0,
    "anotherNumber": 25,
    "exampleString": "value",
    "stringArray": ["a","b", "c"],
    "numberArray": [1,2,3],
    "exampleObject": {
        "anotherString": "anotherValue"
    }
}
```

Granitic uses the term <b>config path</b> to express the fully-qualified name of a variable in configuration. In the
above example `exampleBool` is a config path and so is `exampleObject.anotherString`


## Configuring a Granitic application 

When a Granitic application starts, it looks for configuration files before doing anything else. The location of these 
files is specified using the (`-c`) command line parameter. The default value for `-c` 
is `config`, relative to the working directory from which you start your application, so in previous examples
running:

<pre>
recordstore
</pre>

is the equivalent of running:

<pre>
recordstore -c config
</pre>

The value of the -c parameter is expected to be a comma separated list of:

 * Relative or absolute paths to JSON files
 * Relative or absolute paths to directories
 * Absolute URLs of HTTP or HTTPS resources
 
## Starting your application with a specific config file

Support for directories and remote URIs will be discussed in a later tutorial. For now, open a terminal, navigate to your
`recordstore folder` and run

<pre>
grnc-bind && go build
./recordstore -c config/base.json
</pre>

This runs your recordstore application, specifically stating that a single configuration file 
`resource/config/config.json` should be used. Stop the application with `CTRL+C`

## Injecting configuration into your components

The Go IoC container automatically injects configuration values into your components. Modify the file `artist/get.go` 
so it looks like:

```go
package artist

import (
	"context"
	"github.com/graniticio/granitic/v3/ws"
)

type GetLogic struct{
	EnvLabel string
}

func (gl *GetLogic) Process(ctx context.Context, req *ws.Request, res *ws.Response) {

  an :=  fmt.Sprintf("Hello, %s!", gl.EnvLabel)

  res.Body = Info{
  	Name: an,
  }
}

type Info struct {
	Name string
}

```

We've added a new field to the struct, `EnvLabel` and have changed our _Hello, World!_ message to include the 
value of that field. As the name suggests, this will vary depending on which environment our code is running in.

Now modify the `artistHandler` definition in your `comp-def/common.json` file so it looks like:

```javascript
"artistHandler": {
  "type": "handler.WsHandler",
  "HTTPMethod": "GET",
  "Logic": {
    "type": "artist.GetLogic",
    "EnvLabel": "$environment.label"
  },
  "PathPattern": "^/artist"
}
```

The value starting with the dollar symbol (`$`) is a configuration promise - you are promising your application that Granitic will
find a value in configuration with the _config path_ `environment.label` and inject it into the `artistHandler.Logic` 
component's `EnvLabel` field at application startup.

If you now run:

<pre>grnc-bind && go build && ./recordstore -c config/base.json</pre>

you will see an error message similar to:

<pre>
01/Mar/2019:17:41:16 Z FATAL [grncContainer] No value found at environment.label
</pre>

Granitic adopts a fail-fast model for configuration and will not allow an application to start if it relies on configuration
that is undefined. Rather than just adding the expected configuration to `config/base.json`, we'll
use this opportunity to show how configuration files can be used to make deploying your application in multiple locations 
more straightforward.

### Default values

It is sometimes desirable to be able to define a default value to use if no config exists instead of failing. This can be achieved
with default values. In your `comp-def/common.json` file change the line:

```json
    "EnvLabel": "$environment.label"
```

to

```json
    "EnvLabel": "$environment.label(DEV)"
```

and trying re-running

```
grnc-bind && go build && ./recordstore -c config/base.json
```

and visit [http://localhost:8080/artist](http://localhost:8080/artist) You should see the response:
          
```json
{
"Name": "Hello, DEV!"
}
```

It is **strongly** recommended that you do not store sensitive information such as passwords in default values as default
values are compiled into your application executable.


## Multiple configuration files

Only the simplest applications will use a single configuration file. Complex applications will split their configuration into
multiple files to improve readability and maintainability, but all applications will want to separate the configuration that
is common to each deployment of an application from that configuration that changes across different deployment environments
and from one instance of an application to another.

Generally you would expect only your application's common (or 'base') configuration to be checked into source control with 
environment specific configuration being generated dynamically by your build or configuration management system,

The rest of this tutorial simulates deploying an instance of a web-service across multiple environments and then multiple instances
running on a single server.

Change your `config/base.json` file so it looks like:

```json
{
  "Facilities": {
   "HTTPServer": true,
   "JSONWs": true
  },
  "environment": {
    "label": "TEST"
  }
}
```

Then create a new file:

`/tmp/prod.json`

```json
{
  "environment": {
    "label": "PROD"
  }
}
```

## Configuration merging

We now have a value for the _config path_ `environment.label' defined in three places.

Default value: `DEV` 

`config/base.json`: `TEST`

`/tmp/prod.json`: `PROD`


As you've only changed configuration files, you don't need to rebuild. You can just run:

<pre>
./recordstore -c config,/tmp/prod.json
</pre>


And visit [http://localhost:8080/artist](http://localhost:8080/artist)

You'll see the message has changed to `Hello, PROD!`. This is due to Granitic's process of _configuration merging_ where
multiple configuration files are merged together to provide a single view of application configuration.

The order in which you specify your configuration files when starting your application is important. Any config path that 
is defined in multiple files is replaced with the _rightmost_ definition. Try swapping the order of the configuration 
files and see what happens.
  

## Using configuration to run multiple instances

To maximise use of resources, you may want to run multiple instances of a web service on a single host. This means each 
instance must have a different HTTP port assigned to it. 

Create two new files:

`/tmp/instance-1.json`

```javascript
{
  "HTTPServer": {
    "Port": 8081
  }
}
```

`/tmp/instance-2.json`

```javascript
{
  "HTTPServer": {
    "Port": 8082
  }
}
```

You can now run:

<pre>
./recordstore -c config,/tmp/instance-1.json
</pre>

and in a separate terminal, navigate to the same folder and run:

<pre>
./recordstore -c config,/tmp/instance-2.json
</pre>

And you now have two separate instances of your `recordstore` application running and listening on different ports.


## Facility configuration

In previous examples, you will have noticed that the default HTTP port for Granitic applications is 8080. This is not
hard-coded, it is defined in another configuration file that is included with Granitic itself called a 
`facility configuration file`, which can be found under the `facility/config` folder of your Granitic installation.

These files are serialised into your application's executable so you don't need to have Granitic installed on the environment
you are running your application on. During application startup, Granitic merges this serialised view with your application's configuration files . 

Your application's configuration files take precedence over the built-in facility configuration, so in this example the 
value of `HTTPServer.Port` in `facility/config/httpserver.json`  is replaced with the value in your `/tmp/instance-1.json` or `/tmp/instance-1.json` file.

### Merging rules

The rules by which configuration two files are merged together are specified in the [Granitic GoDoc](https://godoc.org/github.com/graniticio/granitic/config), 
but the following example illustrates the key rules (note the configuration items are an illustration and do not relate to any specific Granitic features)

`a.json`

```javascript
{
    "server": {
        "name": "localhost",
        "network": {
            "interfaces": ["192.168.0.2","127.0.0.1"],
            "sslOnly": false,
            "seed": 1.98311
        },
        "security":{
            "mode": 0
        }
    }
}
```

`b.json`

```javascript
{
    "server": {
        "name": "testserver",
        "network": {
            "interfaces": ["10.123.0.5"],
            "certPath": "/tmp/cert.key",
            "sslOnly": true
        },
        "metrics":{
            "enabled": true
        }
    }
}
```

merged together becomes:

```javascript
{
    "server": {
        "name": "testserver",
        "network": {
            "interfaces": ["10.123.0.5"],
            "certPath": "/tmp/cert.key",
            "sslOnly" true
        },
        "security":{
             "mode": 0
        },
        "metrics":{
            "enabled": true
        }
    }
}
```

Files are merged from left to right, the final value of a configuration item present in more than one file is the one
defined in the rightmost file. The behaviour that might be most unexpected is how arrays are handled - the contents of
arrays are not merged together, but replaced.

## Recap

 * Granitic applications store configuration in JSON files which are loaded when an application starts.
 * Use multiple configuration files to organise your application and to support multiple environments and deployments.
 * Configuration is injected into your Go objects by using config promises in your application's component definition file.
  you change your component definitions.
 * All of your configuration files are merged together with Granitic's built-in facility configuration files to provide
  a single view of configuration.
 
## Next

The next tutorial covers application [logging](003-logging.md) 