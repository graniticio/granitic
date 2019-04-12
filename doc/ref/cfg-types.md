# Configuration type handling

JSON supports a very simple [type model](http://json.org) of strings, booleans, numbers, objects (maps) and arrays of those types.

Your application will use [configuration promises](ioc-definition-files.md) to map a configuration path in your configuration
to a typed field on a Go struct.

Generally the behaviour of mapping between JSON and Go types is as you would expect and as in line with how Go's [JSON 
decoder works](https://blog.golang.org/json-and-go) but there are a few quirks to be aware of.

## Number sizes and signing

You can use a JSON number to populate any Go numeric type, size or signed-ness but information will be silently lost if
the target field's type cannot accommodate the value specified in the JSON number.

As a result, it is recommended that you use `int64` and `float64` types on fields that received numbers from configuration
and then, if necessary validate the value provided and convert to a more specific Go type in your component's 
[StartComponent()](ioc-lifecycle.md) method.


## Mixed types in arrays

If a JSON array contains a mixture of JSON types (e.g. not every element is of the same type), the receiving field
_must_ be of type `interface{}[]`. 

## JSON objects

If your configuration path points to a JSON object/map, the receiving type must either be a struct or a 
`map[string]interface{}`

---
**Next**: [Configuration merging](cfg-merging.md)

**Prev**: [Configuration files](cfg-files.md)