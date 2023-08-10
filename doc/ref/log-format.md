# Log message formatting and location
[Reference](README.md) | [Logging](log-index.md)

---

## Console logging to STDOUT

By default, Granitic performs 'console' logging, using Go's [fmt](https://golang.org/pkg/fmt/) functions
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
      "LogPath": "/tmp/my-app.log",
      "BufferSize": 50
    }
  }
}
``` 

Disables console logging and sends all messages to the file `/tmp/my-app.log`


### Permissions

The OS user that starts your application must have file system permissions sufficient to create and append any log file
you specify. The log file is created with the permissions `0600`

### Buffering

Writing to logs files is buffered and asynchronous - your code does not block while log lines are being written to a file,
_unless_ the buffer is full. This could happen if your application is very busy or if the log file is located on 
particularly slow storage. The size of this buffer is controlled with the configuration path `LogWriting.File.BufferSize`
(default `50`).

### Log rotation

Granitic does not currently have any built-in log rotation capability. It is compatible with tools like `logrotate` as
long as the tool is configured to _truncate_ files, rather than move and recreate them.

## Log line formatting

By default, every message that is logged will be logged as a semi-structured line of text. prefixed with a string like:

`15/Apr/2019:13:44:23 Z DEBUG [grncQueryManager] `

Which is the time zone, date and time the message was logged (to the nearest second), the severity level of the message
and the name of the component that logged the message. If the message is mutli-line (e.g. contains `\n` characters), 
the prefix will only be logged once.

This prefix is configurable by setting a pattern at the config path `LogWriting.Format.PrefixFormat` this pattern
defaults to:

```
"%{02/Jan/2006:15:04:05 Z0700}t %P [%c] "
```

Formats work similar to `fmt` verb patterns. The available `verbs` are:

| Formatting Verb | Meaning and usage |
| ----- | --- |
| %% | The percent symbol |
| %{?}t | The point in time at which the request was received where ? is a standard Go date/time format string (e.g. `02/Jan/2006:15:04:05 Z0700` ). In UTC or local time according to access log configuration ||
| %L | The log level (`DEBUG`, `ERROR`, etc) at which the message was logged. |
| %l | The first character of the level (e.g. `D` for `DEBUG`, etc) at which the message was logged. |
| %P | The log level (`DEBUG`, `ERROR`, etc) at which the message was logged right-padded with spaces so each label takes up five characters. |
| %c | The name of the component which logged the message. |
| %{?}C | The name of the component which logged the message with a fixed length. If the name of the component is longer than ?, it will be truncated to that length. If it is longer, it will be right-padded with spaces. |
| %{?}X | A value from a context.Context that has been made available to the logger via a component you have written implementing [logging.ContextFilter](https://godoc.org/github.com/graniticio/granitic/v2/logging#ContextFilter) where ? is the key to the value 


## Structured JSON logging

Instead of logging each log entry as a line of semi-structured text, Granitic can be configured to use JSON instead.

Setting

```json
{
  "LogWriting": {
    "Format":{
      "Entry": JSON
    }
  }
}
```
will cause messages to be logged like:

```
{"Level":"INFO","Message":"Listening on 9099","Source":"grncCtlServer","Timestamp":"27/Apr/2020:12:14:07 Z"}
```

The name of each generated JSON field and the associated value can be changed in configuration. The default is:

```json
{
  "LogWriting": {
    "Format":{
      "JSON": {
        "Fields": [
          ["Timestamp", "TIMESTAMP", "02/Jan/2006:15:04:05 Z0700"],
          ["Level", "LEVEL"],
          ["Source", "COMPONENT_NAME"],
          ["Message", "MESSAGE"]
        ]
      }
    } 
  }
}
```

Each entry in the ```Fields``` array is a string array with two or three elements. 
  * The name of the field as it will be written in JSON.
  * The type of data (content type) that should be used to set a value for the field
  * An optional argument if required by the content type
  
The available content types are:

| Content Type | Meaning and usage | Argument |
| ----- | --- | --- |
| MESSAGE | The complete message associated with a log entry | N/A |
| FIRST_LINE | The message associated with a log entry up to the first \n char | N/A |
| SKIP_FIRST | The message associated with log entry but strips the first line | N/A |
| COMPONENT_NAME | The full name of the component that logged the message | N/A |
| TIMESTAMP | The datetime at which the message was logged | Any valid Go datetime format string |
| TEXT | The static text specified in the argument | The text to use as a value |
| LEVEL | The full label (DEBUG, ERROR etc) of the logging level at which the entry was logged | N/A |
| CONTEXT_VALUE | A value from a `context.Context` that has been exposed by a [logging.ContextFilter](https://godoc.org/github.com/graniticio/granitic/v2/logging#ContextFilter) component | The string key of the exposed value
| INSTANCE_ID | The ID [assigned to the current instance of your application](adm-instance.md) | N/A |
| LEVEL_MAP | A log level, optionally mapped to a different value. | Comma separated mapping betwen a Granitic log level and some other label (e.g. ```INFO:ANNOUNCE,TRACE:DETAIL```)

### Line prefix and suffix

Each line of JSON formatted log entry can be prefixed or suffixed with a static string by setting `LogWriting.Format.JSON.Prefix`
(default empty string) or `LogWriting.Format.JSON.Prefix` (default ```\n```)

## UTC

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

