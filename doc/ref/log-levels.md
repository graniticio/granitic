# Log levels
Back to: [Reference](README.md) | [Logging](log-index.md)

---
Controlling which components' messages are output to logs and at which severity level is controlled through the
concept of log levels.

The default configuration for Granitic's [logging facility](fac-logger.md) includes this:

```json
"FrameworkLogger":{
  "GlobalLogLevel": "INFO"
},

"ApplicationLogger":{
  "GlobalLogLevel": "INFO"
}
```

This means that, by default, both framework components (Granitic built-in code) and application components (your code)
will log messages that have a severity of `INFO` or higher (e.g `INFO`, `WARN`, `ERROR` and `FATAL`).

This is referred to as the _global log level_, which applies unless you have provided a specific log level for a 
component.

## Overriding the global log level

Like all Granitic facility configuration, your application can change behaviour by overriding that that configuration
in one of your [application's configuration files](cfg-files.md). For example, if you include the following in one
of your configuration files:

```json
"FrameworkLogger":{
  "GlobalLogLevel": "FATAL"
},

"ApplicationLogger":{
  "GlobalLogLevel": "TRACE"
}
```

Granitic's own components will be effectively silenced (only logging 'FATAL', which would indicate a crash) and your
own components would be logging 'TRACE' and above (effectively all messages)