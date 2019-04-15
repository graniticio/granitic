# Log message formatting and location
Back to: [Reference](README.md) | [Logging](log-index.md)

---

## Console logging to STDOUT

By default, Granitic performs 'console' logging, which is to say it use's Go's [fmt](https://golang.org/pkg/fmt/) functions
to print to `os.Stdout`.

If you are starting your application from a terminal you will see log messages appear in that terminal as they are raised.
On Unix style OSes you can redirect output from your application using the standard `>` and `|` operators.

### Disabling STDOUT logging

You can stop your application from sending messages to STDOUT  by setting the following configuration:

```json
{
  "LogWriting": {
    "EnableConsoleLogging": false
  }
}
```

If you want your application to be totally silent, it is recommend you also pass the command line argument `-l FATAL` to
prevent the framework logging messages to the console before your configuration is loaded.


## Logging to a file

You can ask Granitic to send all loggable messages to a standard OS text file by setting:

```json
{
  "LogWriting": {
    "EnableFileLogging": true
  }
}
```

By default, this will cause a file called `granitic.log` to be created in the current working directory (e.g. the 
directory from which you started your application). You can change the name of the file to a relative or absolute
path of your choosing. For example:

```json
{
  "LogWriting": {
    "EnableConsoleLogging": false,
    "EnableFileLogging": true,
    "File":{
      "LogPath": "/tmp/my-app.log"
    }
  }
}
``` 

Disables console logging and sends all messages to the file `/tmp/my-app.log`


### Permissions

The OS user that starts your application must have file system permissions sufficient to create and append any log file
you specify. The log file is created with the permissions `0600`

### Log rotation

Granitic does not currently have any built-in log rotation capability. It _should_ be compatible with third-party
log rotation tools, but this is not fully tested.

## Log message prefixes

By default, every message that is logged will be prefixed with a string like:

`15/Apr/2019:13:44:23 Z DEBUG [grncQueryManager] `

Which is the time zone, date and time the message was logged (to the nearest second), the severity level of the message
and the name of the component that logged the message. If the message is mutli-line (e.g. contains `\n` characters), 
the prefix will only be logged once.

This prefix is configurable by setting a pattern at the config path `LogWriting.Format.PrefixFormat` this pattern
defaults to:

```
"%{02/Jan/2006:15:04:05 Z0700}t %P [%c] "
```

Formats work similar to `fmt` verb patterns. The available `verbs` are documented in the [logging facility](fac-logger.md)
section of this manual.


### UTC

By default, the date and time at which a message is logged is converted to `UTC` before the prefix is printed. To log
local times set:

```json
{
  "LogWriting": {
    "Format":{
      "UtcTimes": false
    }
  }
}
```

---
**Next**: [Web services](ws-index.md)

**Prev**: [Controlling log levels](log-levels.md)

