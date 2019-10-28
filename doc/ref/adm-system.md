# System configuration

The file `facility/config/system.json` contains settings that affect the low-level behaviour of the Granitic framework.
The content of this file is as follows:

```json
{
  "System": {
    "BlockIntervalMS": 2000,
    "BlockRetries": 15,
    "BlockTriesBeforeWarn": 0,
    "FlushMergedConfig": true,
    "GCAfterConfigure": true,
    "GCAfterStart": false,
    "StopIntervalMS": 2000,
    "StopRetries": 15,
    "StopTriesBeforeWarn": 3
  }
}
```

These can values can be overridden in your application configuration.

## Start-up blocking

`System.BlockIntervalMS`, `System.BlockRetries` and `System.BlockTriesBeforeWarn` affect how Granitic handles components that are able to
prevent the application starting until they are ready. This is explained in the [IOC lifecycle](ioc-lifecycle.md) documentation.

`System.BlockRetries` controls how many times a component will be asked if it will allow start-up to proceed. If this number of attempts
is exceeded, your application will shutdown with an error.

`System.BlockIntervalMS` defines the interval between the checks in milliseconds. By default this is `2000`, so Granitic will wait 
for two seconds between a failed ready-to-proceed check and the next check.

`System.BlockTriesBeforeWarn` defines how many failed ready-to-proceed checks are allowed before Granitic logs a warning indicating
application startup is blocked.

## Post-start cleanup

Some aspects of application start-up are relatively memory intensive and Granitic offers a limited set of configuration
options to manage this.

`System.FlushMergedConfig` is by default set to `true` which means the merged, in-memory view of application configuration
is deleted after your application starts. If you need access to this, set this to `false`.

`System.GCAfterConfigure` is also by default set to `true`, which means that Granitic will explicitly invoke [garbage collection](https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html)
after your components have been configured. If `FlushMergedConfig` is set to `true`, this means that the memory used by
the merged configuration will be recovered.

`System.GCAfterStart` is set to `false` by default, but setting this to `true` will make Granitic perform an explicit garbage 
collection after all components have started. This should only be used if you have components that consume a large amount
of recoverable memory during start-up.

## Shutdown blocking

Components that [need to be stopped gracefully](ioc-lifecycle.md) are given a fixed number of 
chances and amount of time to complete their shutdown activities before the application exits. 

`System.StopRetries` controls how many times a component will be asked if it is ready for application exit. If this number of attempts
is exceeded, your application will shutdown immediately.

`System.StopIntervalMS` defines the interval between the checks in milliseconds. By default this is `2000`, so Granitic will wait 
for two seconds between a failed ready-to-stop check and the next check.

`System.StopTriesBeforeWarn` defines how many failed ready-to-stop checks are allowed before Granitic logs a warning indicating
application shutdown is being held up.

---
**Prev**: [Instance identification](adm-instance.md)
