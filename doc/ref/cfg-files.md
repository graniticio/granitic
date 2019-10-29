# Configuration files
[Reference](README.md) | [Configuration](cfg-index.md)

---
Granitic applications require one or more configuration files to be made available when the application starts. Those
files may be present on an accessible filesystem or via an HTTP/HTTPS URL.

## Specifying the files to use

By default, Granitic applications expect to find configuration in folder in the current working directory called `config`.
In most environments it is more common to want to specify the files to use explicitly. This is achieved by using the 
`-c` command line argument to pass a comma separated list of:

  * Relative or absolute paths to JSON files on a filesystem
  * Relative or absolute paths to a filesystem folder containing one or more JSON files
  * Absolute HTTP or HTTPS URLs (including scheme) that return JSON in the response body
  
The order in which these configuration sources are specified are [significant to configuration merging](cfg-merging.md).

### Symlinks

Go's support for symlinks is inconsistent. It is strongly recommended you do not use them when providing paths to files
and folders.

### Folder recursion

When given a folder that may contain JSON files, Granitic performs depth first recursion into any sub folders. Files
and folders are processed in lexicographical order so given the `-c` argument:

`-c conf-dir`

And the folder structure:

```
conf-dir/
  001.json
  a/
   003.json
   b/
     002.json
  z.json   
```
  
Granitic will load configuration files in this order:

```
  conf-dir/001.json
  conf-dir/a/003.json
  conf-dir/b/002.json
  conf-dir/z.json
```

### Name and encoding

JSON configuration files must end with the case-sensitive extension `.json` and be `UTF-8` encoded.

### HTTP URL responses

HTTP URLs invoked to provide JSON configuration must return a response with:

  * A content type of `application/json`
  * A JSON formatted response body
  * A status code of `200`
  
If any of these conditions are not met, your application will fail to start.


## File contents

Any file that is a valid [RFC-7159](https://tools.ietf.org/html/rfc7159) JSON file and is parseable by Go's native JSON
parser is a valid configuration file.

There are no restrictions on file size, formatting or the naming of fields. There is a _weak_ convention that user configuration
files use camel case for field names to distinguish them from the Pascal case used in Granitic's built-in configuration,
but this is entirely optional.


---
**Next**: [Configuration type handling](cfg-types.md)

**Prev**: [Configuration principles](cfg-principles.md)