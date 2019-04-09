# Logging principles

## Logging as a facility

Logging is provided by Granitic as a [facility](fac-logger.md), which means it is optional but enabled by 
default. The [facility documentation](fac-logger.md) explains the default configuration and
what configuration options are available to your application.

## Component-oriented logging and runtime control

Granitic provides a logging framework that allows the log output of components to be controlled
individually, even at runtime. This means, for examples, that debugging information for a component
could be temporarily enabled in production while a service is still running, or that unexpectedly
noisy components can be silenced without having to redeploy or even stop a service.


## Automatically injected loggers

As long as the your component's struct declares a field 

```go
  Log logging.Logger
```

Granitic will inject a configured logger into your component with no further configuration required.


## Standard severity levels

Granitic follows a standard model of labelling log messages according to severity/importance: `TRACE`, 
`DEBUG`,`INFO`, `WARN`, `ERROR`, `FATAL`.

## Control over framework components

Granitic's internal components use the same logging framework and patterns as your application components,
meaning you can enable debugging output for Granitic internals if you need to trace a problem.


---
**Next**: [Adding logging to your code](log-code.md)

**Prev**: [Logging index](log-index.md)