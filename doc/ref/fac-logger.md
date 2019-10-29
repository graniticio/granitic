# FrameworkLogging and ApplicationLogging
[Reference](README.md) | [Facilities](fac-index.md)

---

The FrameworkLogging and ApplicationLogging facilities control how your components, and Granitic's built-in 
components, log messages by importance/severity.

Logging in Granitic is discussed in detail in a [dedicated section of the reference manual](log-index.md) and as such 
this page is limited to explaining the configuration options available.

## Enabling

FrameworkLogging and ApplicationLogging are _enabled_ by default. They can be disabled by setting `Facilities.FrameworkLogging` 
and/or `Facilities.ApplicationLogging` to `false` in your configuration.

Disabling logging causes Granitic to inject 'null' loggers into code that expects an instance of 
[logging.Logger](https://godoc.org/github.com/graniticio/granitic/logging#Logger). This means your code does not need
to worry if logging is disabled or not.

## Configuration

The default configuration for this facility can be found in the Granitic source under `facility/config/logging.json`
and is: 

```json
{
  "LogWriting": {
    "EnableConsoleLogging": true,
    "EnableFileLogging": false,
    "File": {
      "LogPath": "./granitic.log",
      "BufferSize": 50
    },
    "Format": {
      "UtcTimes":     true,
      "Unset": "-"
    }
  },

  "FrameworkLogger":{
     "GlobalLogLevel": "INFO"
  },

  "ApplicationLogger":{
     "GlobalLogLevel": "INFO"
  }
}
```

## Controlling global and per-component log-levels

This is explained in the [log-levels section](log-levels.md) section of the reference manual.

## Controlling the location and format of log files

This is explained in the [formatting and location](log-format.md) section of the reference manual.

## Component reference

The following components are created when both facilities are enabled

| Name | Type |
| ---- | ---- |
| grncApplicationLoggingManager | [logging.ComponentLoggerManager](https://godoc.org/github.com/graniticio/granitic/logging#ComponentLoggerManager) |
| grncFrameworkLoggingManager | [logging.ComponentLoggerManager](https://godoc.org/github.com/graniticio/granitic/logging#ComponentLoggerManager) |

---
**Next**: [JSON Web Services](fac-json-ws.md)

**Prev**: [HTTP Server](fac-http-server.md)