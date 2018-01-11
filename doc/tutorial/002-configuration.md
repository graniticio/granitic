# Tutorial - Configuration

## What you'll learn

 1. How Granitic defines configuration
 2. How to define different configuration files for different environments
 3. How JSON configuration files are merged together

 
## Prerequisites

 1. Follow the Granitic [installation instructions](https://github.com/graniticio/granitic/doc/installation.md)
 2. Read the [before you start](000-before-you-start.md) tutorial
 3. Either have completed [tutorial 1](001-fundamentals.md) or open a terminal and run

<pre>
cd $GOPATH/src/github.com/graniticio
git clone https://github.com/graniticio/granitic-examples.git
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 2
</pre>

## Related GoDoc

https://godoc.org/github.com/graniticio/granitic/config

## Configuration

Granitic applications use JSON files to store configuration which is loaded when the application starts. Any valid JSON file is a valid configuration file:


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

Granitic uses the term <b>configuration path</b> to express the fully-qualified name of a variable in configuration. In the
above example <code>exampleBool</code> is a configuration path and so is <code>exampleObject.anotherString</code>


## Configuring a Granitic application 

When a Granitic application starts, it looks for configuration files before doing anything else. The location of these files is specified using the
(<code>-c</code>) command line parameter. The default value for <code>-c</code> is <code>resource/config</code>, so in previous examples
running:

<pre>
recordstore
</pre>

is the equivalent of running:

<pre>
recordstore -c resource/config
</pre>

The value of the -c parameter is expected to be a comma separated list of:

 * Relative or absolute paths to JSON files
 * Relative or absolute paths to directories
 * Absolute URIs of HTTP or HTTPS resources
 
Support for directories and remote URIs will be discussed in a later tutorial. For now, open a terminal and run:

<pre>
cd $GOPATH/src/granitic-tutorial/recordstore
grnc-bind
go build
./recordstore -c resource/config/config.json
</pre>

This runs your recordstore application, specifically stating that a single configuration file 
<code>resource/config/config.json</code> should be used. Stop the application with <code>CTRL+C</code>

## Injecting configuration into your components

The Go IoC container automatically injects configuration values into your components. Modify the file <code>endpoint/artist.go</code> so it looks like:

```go
package endpoint

import (
  "github.com/graniticio/granitic/ws"
  "context"
)

type ArtistLogic struct {
  EnvLabel string
}

func (al *ArtistLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

  a := new(ArtistDetail)
  a.Name = "Hello, World from " + al.EnvLabel

  res.Body = a
}

type ArtistDetail struct {
  Name string
}
```

We've added a new field to the struct, <code>EnvLabel</code> and have changed our _Hello, World!_ message to include the 
value of that field. As the name suggests, this will vary depending on which environment our code is running in.

Now modify the <code>artistLogic</code> definition in your <code>resource/components/components.json</code> file so it looks like:

```javascript
"artistLogic": {
    "type": "endpoint.ArtistLogic",
    "EnvLabel": "conf:environment.label"
}
```

This is a configuration promise - you are telling Granitic to find a value in configuration with the _configuration path_ 
<code>environment.label</code> and inject it into the
<code>artistLogic</code> component's <code>EnvLabel</code> field at runtime.

If you now run:

<pre>grnc-bind && go build && ./recordstore -c resource/config/config.json</pre>

you will see an error message similar to:

<pre>
17/Jan/2017:16:39:42 Z FATAL [grncContainer] No value found at environment.label
</pre>

Granitic adopts a fail-fast model for configuration and will not allow an application to start if it relies on configuration
that is undefined. Rather than just adding the expected configuration to <code>resource/config/config.json</code>, we'll
use this opportunity to show how configuration files can be used to make deploying your application in multiple locations 
more straightfoward.


## Multiple configuration files

Only the simplest applications will use a single configuration file. Complex applications will split their configuration into
multiple files to improve readability and maintainability, but most applications will want to separate the configuration that
is common to each deployment of an application from that configuration that changes across different deployment environments
and from one instance of an application to another.

The rest of this tutorial simulates deploying an instance of a web-service across multiple environments and then multiple instances
running on a single server.

Create two new files:

<code>resource/env/production.json</code>

```json
{
  "environment": {
    "label": "PROD"
  }
}
```


<code>resource/env/development.json</code>

```json
{
  "environment": {
    "label": "DEV"
  }
}
```

As you've only changed configuration files, you don't need to rebuild. You can just run:

<pre>
./recordstore -c resource/config,resource/env/production.json
</pre>

or
 
<pre>
./recordstore -c resource/config,resource/env/development.json
</pre>

And visit [http://localhost:8080/artist](http://localhost:8080/artist) to see a different result depending on which config file you're using.


### Tidying up

In later tutorials, we'll want the option of running <code>recordstore</code> without specifiying a list of config files, so change
your <pre>resource/config/config.json</pre> file so that it looks like:

```json
{
    "Facilities": {
        "HttpServer": true,
        "JsonWs": true
    },
    "environment": {
        "label": "UNSET"
    }
}

```


## Overriding configuration

To maximise use of resources, you may want to run multiple instances of a web service on a single host. This means each instance must have a different HTTP port assigned to it. 

This is an example of how you can override previously defined configuration items with new values.

Create two new files:

<code>resource/instance/instance-1.json</code>

```javascript
{
  "HttpServer": {
    "Port": 8081
  }
}
```

<code>resource/instance/instance-2.json</code>

```javascript
{
  "HttpServer": {
    "Port": 8082
  }
}
```

You can now run:

<pre>
./recordstore -c resource/config,resource/env/development.json,resource/instance/instance-1.json
</pre>

and in a separate terminal run:

<pre>
cd $GOPATH/src/granitic-tutorial/recordstore
./recordstore -c resource/config,resource/env/development.json,resource/instance/instance-2.json
</pre>

And you now have two separate instances of your recordstore application running and listening on different ports.


## Configuration merging

In previous examples, you will have noticed that the default HTTP port for Granitic applications is 8080. This is not
hard-coded, it is defined in another configuration file that is included with Granitic itself called a 
<code>facility configuration file</code>. Making sure your applications can read these built-in configuration files is why you 
need to set the <code>GRANITIC_HOME</code> environment variable to point to an installation of Granitic.

During startup, Granitic builds up a single view of configuration by merging together all of your configuration files with all of the 
built-in configuration files that can be found under <code>$GRANITIC_HOME/resource/facility-config</code>.

You can see the order in which Granitic is merging configuration files together by starting an application with the <code>-l TRACE</code>
parameter, which sets Granitic's initial log-level to <code>TRACE</code>. Your application's configuration files take precedence over
the built-in facility configuration files, so in this example the value of <code>HttpServer.Port</code> in <code>facility-config/httpserver.json</code> is replaced
with the value in your <code>recordstore-1.json</code> or <code>recordstore-2.json</code> file.

### Merging rules

The rules by which configuration two files are merged together are specified in the [Granitic GoDoc](https://godoc.org/github.com/graniticio/granitic/config), but the following example 
illustrates the key rules (note the configuration items are an illustration and do not relate to any specific Granitic features)

<code>a.json</code>

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

<code>b.json</code>

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