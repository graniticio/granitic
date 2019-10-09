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

### Instrumentation

The HTTP server supports and coordinates the [instrumentation of web service requests](ws-instrumentation.md) automatically
finding a component you have registered that implements [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager).

There are two configuration settings that affect this behaviour. 

#### Auto-wiring

Setting `HTTPServer.DisableInstrumentationAutoWire` to `true`
means that the HTTP server will not automatically look for a component that implements [instrument.RequestInstrumentationManager](https://godoc.org/github.com/graniticio/granitic/instrument#RequestInstrumentationManager).
Instead you will need to explicit provide an instrumentation manager via a framework modifier like:

```json
"frameworkModifiers": {
  "grncHTTPServer": {
    "InstrumentationManager": "myInstrumentationManager"
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