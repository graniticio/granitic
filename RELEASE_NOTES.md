# Granitic 2.1 Release Notes

Granitic 2.1 adds features and fixes that are backwards-compatible with code based on Granitic 2.0.x 

See the [milestone on GitHub](https://github.com/graniticio/granitic/issues?utf8=%E2%9C%93&q=is%3Aissue+milestone%3Av2.1.0+)
for a complete list of issues resolved in this release.

## Access Logging

The access logging feature of the HTTPServer facility has the following changed behaviour.

  * Improved stopping behaviour (silently ignores LogRequest when application is stopping) and closes log line channel correctly
  * Now explicitly switches to synchronous/blocking logging when `HTTPServer.AccessLog.LineBufferSize` is set to zero or less. Default is 10.
  * Data stored in a `context.Context` can now be logged in access log lines (using the `%{KEY}X` verb) as long
  as you have created a component that implements [logging.ContextFilter](https://godoc.org/github.com/graniticio/granitic/logging#ContextFilter)
  
## Application logging

Data stored in a `context.Context` can now be logged in the prefix of applicaiton log lines (using the `%{KEY}X` verb) as long
as you have created a component that implements [logging.ContextFilter](https://godoc.org/github.com/graniticio/granitic/logging#ContextFilter)

## HTTP Response Status Codes

The default set of HTTP status codes (described [here](https://granitic.io/ref/error-handling)) can now be overridden in
configuration [as described here](https://granitic.io/ref/json-web-services)


## Request IDs

If [request identification](https://granitic.io/ref/request-identity) has been set up, your code can now find the string
ID for the current request by calling `ws.RequestID(context.Context)`