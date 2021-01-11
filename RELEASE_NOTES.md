# Granitic 2.3 Release Notes

Granitic 2.3 adds features and fixes that are backwards-compatible with application code based on 
Granitic 2.x 

## Request Instrumentation

  * A new function `instrument.Amend(interface{})` allows your application code to provide additional
  data to your `instrument.Instrumentor` after an event has been started.
  * Granitic now calls your `Instrumentor.Amend` function with the flag:
     * `UserIdentity` and an instance of `iam.ClientIdentity` after the user has been positively identified or tagged as anonymous.
     * `Request` and an instance of `*ws.Request` as soon as the `ws.Request` object is created. Implementations should be cautious when using this object as fields are not guaranteed to be set.
     * `ResponseWriter` and an instance of `*httpendpoint.HTTPResponseWriter' after the HTTP response is written.
  * Instrumentation can be disabled on a per-handler basis. In your `WsHandler's` configuration set `DisableInstrumentation`
  to `true`
  
  
