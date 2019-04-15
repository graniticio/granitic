# Adding logging to your code
Back to: [Reference](README.md) | [Logging](log-index.md)

---
## Having Granitic inject a logger

Your code can send log messages via the [logging.Log interface](https://godoc.org/github.com/graniticio/granitic/logging#Logger)
if the struct that provides the implementation for your component has a field with exactly this name and
type:

```go
  Log logging.Logger
```

Granitic will automatically inject a configured and ready to use logger into your component.

## Logging messages

Your code can send messages to be logged using one of the `LogXXXX` methods on 
[logging.Log](https://godoc.org/github.com/graniticio/granitic/logging#Logger). Convenience methods
are provided to log messages at each of the standard logging severity levels.

For example, to log a message at the `INFO` level, a method on your component might look like

```go
func (mt *MyType) doSomething() {
  mt.Log.LogInfof("An informational message")
}
```

This message will only actually be written to [stdout/a log file](log-runtime.md) if the log level configuration
of the logger allows this (see below).

### Context methods

Each of the convenience methods on `logging.Log` have a variant to allow a `context.Context` to be provided.
Using these methods will allow you to include information held in the context in the meta-data 
[printed along with each log line](log-format.md)

## Formatting messages

The `f` suffix on each method indicates you can use the same verbs pattern as used by Go's 
[fmt](https://golang.org/pkg/fmt/) package.

For example:

```go
  LogInfof("Running job %d of %d", current, total)
```

## Stack traces

When recovering from a panic it is often useful to record the [stack trace](https://golang.org/pkg/runtime/#Stack)
of function/method calls leading to the panic. The logger methods `LogErrorfWithTrace` and `LogFatalfWithTrace`
generate a stack trace and append it to the message to be logged.

## Checking to see if a message will be logged

Sometimes it is relatively expensive to construct a message to be logged. In these cases it is good
practise to see if the message _would_ be logged. The can be achieved with the `IsLevelEnabled` method
on the logger.

For example:

```go
if i.logger.IsLevelEnabled(logging.Debug) {
  i.logger.LogDebugf("Loading configuration from: ")
  
  for _, fileName := range configPaths {
    i.logger.LogDebugf(fileName)
  }
}
```

---
**Next**: [Log levels](log-levels.md)

**Prev**: [Logging principles](log-principles.md)