# Granitic 2.1 Release Notes

Granitic 2.1 adds features and fixes that are backwards-compatible with code based on Granitic 2.0.x 

See the [milestone on GitHub](https://github.com/graniticio/granitic/issues?utf8=%E2%9C%93&q=is%3Aissue+milestone%3Av2.1.0+)
for a complete list of issues resolved in this release.

## Access Logging

The access logging feature of the HTTPServer facility has the following changed behaviour.

  * Improved stopping behaviour (silently ignores LogRequest when application is stopping) and closes log line channel correctly
  * Now explicitly switches to synchronous logging when `HTTPServer.LineBufferSize` is set to zero or less
  * 