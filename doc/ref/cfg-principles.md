# Configuration principles

## Static vs dynamic configuration

Any value that modifies how your application code works is considered to be configuration. A small portion of this
configuration is considered to be 'static' configuration which is effectively part of your code. For example, the configuration
that tells a [handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) whether it should
listen for `GET` or `POST` requests is fundamental and will never change.

The vast majority of configuration is dynamic configuration that _could_ or _will_ change depending on the environment
that the code is running in (hostnames, ports etc) or that might be beneficial to change without rebuilding your
application (error messages, tuning parameters etc).

## Runtime configuration

Granitic is designed to defer configuration of your application until application startup time. Your application
will be able to load configuration files from any accessible file path or from a remote URL.

This means that it is possible to build applications where the distributable is just the compiled native binary which 
means that you do not different separate _development_ and _production_ builds of your application.

## Configuration layers

Granitic thinks of configuration in terms of layers, where each layer adds to or replaces the the configuration defined
in the previous layer, but doesn't have to re-declare every piece of configuration for each layer. This is fully
explained in the section on [configuration merging](cfg-merging.md), but briefly:

Each application has a _base_ layer of configuration. This layer defines the majority of configuration that is unlikely
to change between environments. According to your preference, this layer may be complete enough to start the application
in a _development_ environment. This layer is likely to be checked into source control as one or more files.

When in development, a developer may define a _local_ layer of configuration that adds or changes enough configuration
to get the application working in their development environment (e.g. use a local database, turn off some production-only
integrations). This layer probably won't be checked into source control.

Once the application is deployed to a named environment (_integration_, _staging_, _production_ etc), a configuration layer
representing the configuration specific to this environment will be made available to the application via files or URLs. 
It is likely that the configuration for this environment will have been _generated_ by a build or configuration management 
system.

Finally, if multiple instances of your application are deployed to a single environment, an _instance_ layer may be required.
This layer differentiates each instance and simply contain a unique name for each instance.

## Configuration paths

Granitic supports configuration files in JSON or YAML. Both these formats define hierarchies of data that can be 
programmatically navigated. The route from the root of a JSON/YAML document to a field containing a value or to a data
structures like a  map or array is called a _configuration path_

For example:

```json
{
  "databases": {
    "application": {
      "host": "localhost",
      "user": "app"
    },
    "monitoring": {
      "host": "some-other-host"
    }
  }
}
```

In this file `databases.application.user`, `databases.application` and `databases.monitoring.host` are configuration paths. 

It is these configuration paths that are defined in your [component definition files](ioc-definition-files.md) to instruct
Granitic how to configure your application at runtime.

---
**Next**: [Configuration files](cfg-files.md)

**Prev**: [Configuration index](cfg-index.md)



