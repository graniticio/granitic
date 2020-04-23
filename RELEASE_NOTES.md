# Granitic 2.2 Release Notes

Granitic 2.2 adds features and fixes that are backwards-compatible with code based on Granitic 2.0.x 

## STDOUT access logging

The HTTP Server access log can now write to `STDOUT` if you set `HTTPServer.AccessLog.LogPath` to `STDOUT` in configuration.

## Save merged configuration

Supply the `-m` flag and a path to a file when starting your application will cause Granitic to write it's final, merged
view of your application's configuration to that file (as JSON) and then exit.

This is useful when debugging problems introduced my merging configuration files.