# Granitic 2.2 Release Notes

Granitic 2.2 adds features and fixes that are backwards-compatible with code based on Granitic 2.0.x 

## STDOUT access logging

The HTTP Server access log can now write to `STDOUT` if you set `HTTPServer.AccessLog.LogPath` to `STDOUT` in configuration.