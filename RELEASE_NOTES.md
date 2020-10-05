# Granitic 2.3 Release Notes

Granitic 2.2 adds features and fixes that are backwards-compatible with application code based on 
Granitic 2.0.x 

## Request Instrumentation

  * A new function `instrument.Amend(interface{})` allows your application code to provide additional
  data to your `instrument.Instrumentor` after an event has been started.
  * Granitic now calls your `Instrumentor.Amend` function with the flag `UserIdentity` and an instance of
  `iam.ClientIdentity` after the user has been postively identified or tagged as anonymous.
  * Granitic now calls your `Instrumentor.Amend` function with the flag `Request` and a `*ws.Request` 
  as soon as the `ws.Request` object is created. Implementations should be cautious when using this object as fields are not guaranteed to be set.
  
  
  
