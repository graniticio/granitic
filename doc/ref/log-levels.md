# Log levels
[Reference](README.md) | [Logging](log-index.md)

---
Controlling which components' messages are output to logs and at which severity level is controlled through the
concept of log level thresholds (or just log levels).

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
own components would be logging 'TRACE' and above (effectively all messages).


## Setting a log level for a specific component

You can change the log level for an individual component without affecting the global log level. For example:

```json
"FrameworkLogger":{
  "GlobalLogLevel": "INFO",
  "ComponentLogLevels": {
    "grncQueryManager": "DEBUG",
    "grncHTTPServer": "FATAL"
  }
},

"ApplicationLogger":{
  "GlobalLogLevel": "INFO",
  "ComponentLogLevels": {
    "myComponent": "TRACE"
  }
}
```

This shows the built-in Granitic components [grncQueryManager](fac-query.md) and [grncHTTPServer](fac-http-server.md)
having their logging threshold set to `DEBUG` and `FATAL` respectively. Additionally the application component `myComponent`
has its threshold set to `TRACE`.

### Specific levels take precedence

If a log level is explicitly set, that level will always take precedence over the global log level. So if a component
is explicitly set to `TRACE`, it will log all messages, even if the global log level was set to `FATAL`.

Conversely, a component with a log level set to `FATAL` would only output FATAL messages, even if the global level was
set to `TRACE`.

## Setting log levels at runtime

If you enable Granitic's [runtime control facility](fac-runtime.md), you use the [runtime-ctl](rtc-command.md) command
line tool or the [runtime control API](rtc-api.md) to change the global or component-specific log levels when your
application is running.

The commands available are documented here in the [built-in commands](rtc-built-in.md) section of the [runtime control](rtc-index.md)
documentation.

## Logging during the bootstrap phase

Loading and applying your logging configuration is one of the first things Granitic does, but there is logic that is
run before that point (finding, loading and validating your configuration, for one). This pre-configuration state
is known as the bootstrap phase. You can control logging behaviour during this phase by passing the `-l` argument
to your application and providing one of the standard log levels (from `TRACE` to `FATAL`).

---
**Next**: [Formatting and output location](log-format.md)

**Prev**: [Adding logging to your code](log-code.md)