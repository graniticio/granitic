# Remote configuration
[Reference](README.md) | [Administration](adm-index.md)

---

When starting a Granitic application, you provide a list of folders and configuration files that are [merged](cfg-merging.md)
together to form a single view of application configuration. These files may be stored on a filesystem, but may also
be served by an HTTP server.

For example:

```shell script
myapp -c config/,/var/myapp/local.json,http://example.com/conf
```

URLs may be any valid URL as defined by the [Go URL parser](https://golang.org/pkg/net/url/#Parse).

## Requirements

The response to the request to a config-providing URL must contain:

  * A content type of `application/json`
  * A JSON formatted response body
  * A status code of `200`
  
If any of these conditions are not met, your application will fail to start.

---
**Next**: [Instance identification](adm-instance.md)

**Prev**: [Administration index](adm-index.md)