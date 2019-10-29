# HTTP Server

Enabling the HTTPServer facility makes your [web service endpoints](ws-handlers.md) accessible to callers via HTTP. The 
HTTP server that is created is a wrapper over Go's [built-in HTTP serving functionality](https://golang.org/pkg/net/http/).

It is agnostic of the content types being accepted or served by your handlers - this is controlled by enabling either the [JSONWS](fac-json-ws.md) 
or [XMLWs](fac-xml-ws.md) facility and/or defining custom content type handinling in your [endpoints](ws-handlers.md).

This page covers the following:

  * Enabling and configuring the HTTP server
  * Extending functionality by providing components that implement particular interfaces
  * How the HTTP server is affected by application [lifecycle events](ioc-lifecycle.md)
  * Enabling and configuring access logging


## Enabling

The HTTPServer facility is _disabled_ by default. To enable it, you must set the following in your configuration

```json
{
  "Facilities": {
    "HTTPServer": true
  }
}
```

## Configuration

The default configuration for this facility can be found in the Granitic source under `facility/config/httpserver.json`
and is:

```json
{
  "HTTPServer":{
    "Port": 8080,
    "Address": "",
    "AllowEarlyInstrumentation": false,
    "DisableInstrumentationAutoWire": false,
    "MaxConcurrent": 0,
    "TooBusyStatus": 503,
    "AutoFindHandlers": true,
    "AccessLogging": false,
    "AccessLog": {
      "LogPath": "./access.log",
      "LogLinePreset": "framework",
      "UtcTimes": true,
      "LineBufferSize": 10
    },
    "RequestID": {
      "Enabled": false,
      "Format": "UUIDV4",
      "UUID":{
        "Encoding": "RFC4122"
      }
    }
  }
}
```

### Listening

By default the HTTP server listens on port `8080` on all available IP addresses (including localhost), on IPV4 and IPV6. Setting
`HTTPServer.Port` in your configuration changes the TCP/IP that clients can connect to.

To listen on a specific IP address, you should set `HTTPServer.Address` to an IP address associated with 
one of the network interfaces available to your server.

This example:

```json
{
  "HTTPServer": {
    "Port": 80,
    "Address": "192.168.0.142"  
  }
}
```  

Will start an HTTP server that _only_ listens on `192.168.0.142:80`

#### HTTPS

The Granitic HTTP server does not support HTTPS at this time. Your application should be deployed behind a web server or load-balancer
with HTTPS support if this is required.

### Load management

By default the HTTP server will accept an unlimited number of concurrent requests. This behaviour can be changed
by setting `HTTPServer.MaxConcurrent` to an integer greater than zero.

Any client attempting to connect to the server while it is already handling the maximum concurrent requests will
receive an error response with the HTTP Status code defined in `TooBusyStatus` (deafult `503`).

### Finding endpoints

By default any [component](ioc-principles.md) you have created that implements the [httpendpoint.Provider](https://godoc.org/github.com/graniticio/granitic/httpendpoint#Provider)
interface will automatically found and added to the list of endpoints that can be matched against incoming HTTP
requests.

See the [web services](ws-index.md) documentation for more details on creating handler components. 

This behaviour can disabled by setting `HTTPServer.AutoFindHandlers` to false. This is advanced behaviour only
generally required when you are running multiple custom instances of the Granitic HTTP server in the same application.


## Extending functionality

### Handling abnormal statuses

Granitic's [web service processing pipeline](ws-pipeline.md) delegates the responsibility for writing HTTP responses
to the [web service handlers](ws-handlers.md), or more specifically the [ws.ResponseWriter](https://godoc.org/github.com/graniticio/granitic/ws#ResponseWriter)
attached to the handler.

There many circumstances under which the HTTP server will reject an inbound request before it reaches a handler, most
commonly when the request cannot be matched to a handler (a `404`). In these circumstances, the HTTP server still needs
to be able to construct an HTTP response body that is consistent with 'normal' responses.

If you are using the [JSONWS](fac-json-ws.md) or [XMLWs](fac-xml-ws.md) facility, this is handled automatically. If you
not using either of those facilities (or want different behaviour) you must provide the HTTP server with an
abnormal status writer by creating a component that implements [ws.AbnormalStatusWriter](https://godoc.org/github.com/graniticio/granitic/ws#AbnormalStatusWriter)
and instructing the HTTP server to use it by providing a [framework modifier](ioc-definition-files.md) like:

```json
"frameworkModifiers": {
  "grncHTTPServer": {
    "AbnormalStatusWriter": "myStatusWriter"
  }
}
``` 

where `myStatusWriter` is the name of your component that implements [ws.AbnormalStatusWriter](https://godoc.org/github.com/graniticio/granitic/ws#AbnormalStatusWriter).

### Request identification

If you have created a component that implements [httpserver.IdentifiedRequestContextBuilder](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#IdentifiedRequestContextBuilder)
that component will be automatically be used to create, derive or inherit a ID for the current request using information included in the HTTP request.

See [request identification](ws-identity.md) for more information.

You may instruct Granitic to assign a UUID V4 ID to each request by setting `HTTPServer.RequestID.Enabled` to
`true`. By default these IDs will be standard [RFC4122](https://tools.ietf.org/html/rfc4122) formatted UUIDs, but you
can choose to alter the formatting by setting `HTTPServer.RequestID.UUID.Encoding` to `Base32` or `Base64`
"RFC4122":

### Instrumentation

The HTTP server supports and coordinates the [instrumentation of web service requests](ws-instrumentation.md) automatically
finding a component you have registered that implements [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager).

There are two configuration settings that affect this behaviour. 

#### Auto-wiring

Setting `HTTPServer.DisableInstrumentationAutoWire` to `true`
means that the HTTP server will not automatically look for a component that implements [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager).
Instead you will need to explicit provide an instrumentation manager via a framework modifier like:

```json
{
  "frameworkModifiers": {
    "grncHTTPServer": {
      "InstrumentationManager": "myInstrumentationManager"
    }
  }
}
``` 

#### Early instrumentation

By default Granitic does not start instrumentation until after a request has been accepted (e.g. after checks have taken
place to make sure the server is not too busy or suspended). This is to provide some protection against deliberate or
accidental denial-of-service attacks putting load on your instrumentation implementation at the cost of losing some
fine-grained timing information around the very early stages of request processing and losing all visibility of requests
that were rejected because the server was busy.

If you want to be able to instrument these types of request, set `HTTPServer.AllowEarlyInstrumentation` to `true`.

## Access logging

Granitic can be configured to write a summary of each request received to a log file, similar to most web and application
servers. This feature is disabled by default and can be enabled by setting `HTTPServer.AccessLogging` to true in your
configuration.

This will create (or append) a UTF-8 encoded file called `access.log` in the same directory that you started your application (e.g. your
application's working directory). You can change this specifying a full or relative path to a file with
`HTTPServer.AccessLog.LogPath`. Your application must have filesystem permission to create and edit the file at that path.  

The format and information recorded for each request is highly customisable and is described in detail below.

### Timestamps

You may choose to have the timestamps associated with each request as UTC (recommended and the default), or in the local
time of your server. To use local time, set `HTTPServer.AccessLog.UtcTimes` to `false`.

### Non-blocking writing

Lines are written to the access log asynchronously. You web service calls will not block for log writing as long as there
is space in the log line buffer. The size of this buffer is defined at `HTTPServer.AccessLog.LineBufferSize` with a default 
value of `10`. If your application is writing logs to a slow storage system or handles large numbers of simultaneous 
requests, you might want to adjust this value.

If you want to force blocking writing (useful for tests), set `HTTPServer.AccessLog.LineBufferSize` to zero or less.

## Log line format

The information you want to include in each line of the access log is controlled by a format string comprised of 'verbs'
and fixed characters similar to [fmt.Printf](https://golang.org/pkg/fmt/) or HTTPD. You may choose from a preset
format or define your own.

### Preset formats

Granitic currently specifies threee preset formats. You can set `HTTPServer.AccessLog.LogLinePreset` to `framework`, `common`
or `combined`. The default is `framework`. The `common` and `combined` formats are intended to be similar to the 
[Apache HTTPD log formats](http://httpd.apache.org/docs/current/mod/mod_log_config.html) of the same name.

| Name | Format | Example |
| --- | --- | --- |
| framework | %h XFF[%{X-Forwarded-For}i] %l %u [%{02/Jan/2006:15:04:05 Z0700}t] "%m %U%q" %s %bB %{us}Tμs | [::1]:49574 XFF[-] - - [09/Oct/2019:14:59:55 Z] "POST /test" 200 -B 252μs |
| common | %h %l %u %t "%r" %s %b | [::1]:49580 - - [09/Oct/2019:15:01:35 +0000] "POST /test HTTP/1.1" 200 - |
| combined | %h %l %u %t "%r" %s %b "%{Referer}i" "%{User-agent}i" | [::1]:49584 - - [09/Oct/2019:15:02:37 +0000] "POST /test HTTP/1.1" 200 - "-" "-" |


### Available verbs

| Formatting Verb | Meaning and usage |
| ----- | --- |
| %% | The percent symbol |
| %b | The number of bytes (excluding headers) sent to client or the - symbol if zero |
| %B | The number of bytes (excluding headers) sent to client or the 0 symbol if zero |
| %D | The wall-clock time the service spent processing the request in microseconds |
| %h | The host (as IPV4 or IPV6 address) from which the client is connecting |
| %{?}i | The string value of a header included in the HTTP request where ? is the case insensitive name of the header |
| %l | Prints the - symbol. For compatibility with common log formats always. |
| %m | The HTTP method (GET, POST etc) of the request |
| %q | The query string of the request, including a leading ? character |
| %r | The HTTP request line (method, path and HTTP version) |
| %s | The HTTP status code (200, 404 etc) sent to the client with the response |
| %{?}t | The point in time at which the request was received where ? is a standard Go date/time format string (e.g. `02/Jan/2006:15:04:05 Z0700` ). In UTC or local time according to access log configuration |
| %{?}T | The wall-clock time the service spent processing the request in a unit specified by ? where s gives seconds, ms gives milliseconds and us gives microseconds |
| %u | A string representation of the ID of the user on whose behalf the request is being made. Only available if [IAM is configured](ws-iam.md), otherwise the - symbol is printed |
| %U | The path portion of the HTTP request line |
| %{?}X | A value from a context.Context that has been made available to the access logger via a component you have written implementing [logging.ContextFilter](https://godoc.org/github.com/graniticio/granitic/logging#ContextFilter) where ? is the key to the value 

## Lifecycle

The IOC component that represents the HTTP server is integrated with Granitic's  [component lifecycle model](ioc-lifecycle.md) and
has different behaviour depending on which lifecycle state or transition the component is in.

The server can be suspended and resumed using [runtime control](rtc-index.md)

### Start

 * Finds any other components required to run (handlers, instrumentation, abnormal status writer, request identifier)
 * DOES NOT open or listening on the configured address and port
 
### Allow Access

 * Confirms that it is possible to listen on the configured address and port
 * Starts listening for HTTP requests on the configured address and port
 
### Suspend
 
 * Keeps listening for requests but sends a 'too busy' response (default 503)
 
### Resume

 * Allows requests to be processed again
 
### Prepare to stop
 
 * Keeps processing existing requests but sends a 'too busy' response (default 503) for any new requests
 
### Ready to stop check

 * Returns true if no requests are currently being processed
 
### Stop

 * Shuts down the underlying HTTP server, terminating any requests that are still running

## Component reference

The following components are created when this facility is enabled:

| Name | Type |
| ---- | ---- |
| grncHTTPServer | [httpserver.HTTPServer](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#HTTPServer) |
| grncAccessLogWriter | [httpserver.AccessLogWriter](https://godoc.org/github.com/graniticio/granitic/facility/httpserver#AccessLogWriter) |