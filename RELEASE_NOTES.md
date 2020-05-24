# Granitic 2.2 Release Notes

Granitic 2.2 adds features and fixes that are backwards-compatible with code based on Granitic 2.0.x 

## Logging

### JSON structured access and application logging

Application logs and access logs can now configured to output a user-defined JSON structure for each log event. 
Refer to the reference manual for [application logging](https://granitic.io/ref/logging-format-output) 
and the [HTTP server facility](https://granitic.io/ref/http-server) for more details. 

### STDOUT access logging

The HTTP Server access log can now write to `STDOUT` if you set `HTTPServer.AccessLog.LogPath` to `STDOUT` in configuration.

### Multiple ContextFilters

Your application can now have as more than one component that implements logging.ContextFilter. If more than
one component is found, they will be merged into a [logging.PrioritisedContextFilter](https://godoc.org/github.com/graniticio/granitic/logging#PrioritisedContextFilter)

You can control the how keys present in more than one ContextFilter are resolved to a single value by
having your ContextFilters implement [logging.FilterPriority](https://godoc.org/github.com/graniticio/granitic/logging#FilterPriority)

## Application start up

### Save merged configuration

Supply the `-m` flag and a path to a file when starting your application will cause Granitic to write it's final, merged
view of your application's configuration to that file (as JSON) and then exit.

This is useful when debugging problems introduced my merging configuration files.

### Instance IDs

Setting the `-u` flag when starting your application will cause Granitic to generate a V4 UUID and use that as the ID
for that instance your application.

Your application's instance ID is now available as configuration at the path `System.InstanceID`

## Query Manager

## Slices as parameters

When using the query manager (either directly or via the RDBMS interfaces) to construct a parameterised query, you
may now provide slices/arrays of some types as parameters. The default behaviour is to format the slice as a comma separated
list. Slices may contain any of the types supported as individual parameters - basic int types, booleans, strings and their
[nilable equivalents](https://granitic.io/ref/nilable-types).

## Web Services

### Slices as parameters in query strings and request paths

[Request path and query binding](https://granitic.io/ref/capture-web-service-data) now supports binding into a slice. 
The parameter value should be expressed as a comma-separated list (e.g. `/some/p,a,t,h` or `/path?a=1,2,3&b=true,false`)

Target slices can be a slice of any Go basic type (except `uintptr`, `complex64` and `complex128`) and any 
[nilable type](https://granitic.io/ref/nilable-types).

## Bug fixes

### Query manager default configuration

Enabling the query manager facility with its default configuration was causing a fatal error on application startup.