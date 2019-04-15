# Configuration merging
Back to: [Reference](README.md) | [Configuration](cfg-index.md)

---
A key principle in Granitic's configuration model is that of [configuration layers](cfg-principles.md), where an
application has a base configuration then additional configuration is layered on for specific environments
and deployments of your application.

The is no strong concept of _layer_ in Granitic, instead a layer is one or more configuration files that logically
represent a layer. In a simple application, each layer will probably be represented by a single JSON file, but more
complex applications may choose to have multiple files per each layer.

At runtime, Granitic resolves all of the configuration files and URLs that are available to your application
and merges them together to form a single unified view of application confgiuration.

## Example

This process is  explained in detail, but requires an example, so imagine an application that is started with the following
command:

`exampleapp -c base.json,production,instance.json`

And the following configuration files available:

`base.json`

```json
{
  "instance": "dev-example",
  "ApplicationLogger":{
    "GlobalLogLevel": "DEBUG"
  }
}

```

`production/config.json`

```json
{
  "instance": "prod-example"
}
```

`production/logging.json`

```json
{
  "ApplicationLogger":{
    "GlobalLogLevel": "ERROR"
  }
}
```

`instance.json`

```json
{
  "instance": "example-1/8080"
}

```

## Merging order

The order in which configuration files are presented to Granitic is significant. In the event that more than one 
configuration file specifies a value for a given configuration path, the final merged configuration will contain the
value in the _rightmost_ file in which the configuration path is defined. 

So in the above example, the order in which Granitic will process configuration files is:

  * base.json
  * production/config.json
  * production/logging.json
  * instance.json

After merging is complete, the value of the configuration path:

  * `instance` will be `example-1/8080` (last defined in `instance.json`)
  * `ApplicationLogger.GlobalLogLevel` will be `ERROR` (last defined in`production/logging.json`)


## Merging rules for JSON types

If a configuration path appears in two or more files, the eventual value of that field is determined by the types involved.

### Strings, numbers and booleans

When merging configuration paths that have a boolean, string or number value, the eventual value is simply the value
in the rightmost file.

For example:

`exampleapp -c a.json,b.json`

`a.json`

```json
{
  "a": 1,
  "b": false,
  "c": "apple",
  "d": -10
}
```

`b.json`

```json
{
  "a": 2,
  "b": true,
  "c": "orange" 
}
```

results in an effective configuration of:

```json
{
  "a": 2,
  "b": true,
  "c": "orange",
  "d": -10 
}
```

### Objects (maps)

When merging configuration paths that have a JSON object (map) value, the result is a map containing all referenced keys.
When a key appears in more than one definition of the map, the eventual value is that from the rightmost file.

For example:

`exampleapp -c a.json,b.json`

`a.json`

```json
{
  "someObject": {
    "a": 1,
    "b": false,
    "c": "apple",
    "d": -10  
  }
}
```

`b.json`

```json
{
  "someObject": {
    "a": 2,
    "b": true,
    "c": "orange" 
  }
}
```

results in an effective configuration of:

```json
{
  "someObject": {
    "a": 2,
    "b": true,
    "c": "orange",
    "d": -10,
  } 
}
```

### Arrays

When merging configuration paths that have an array of any type, the eventual value is the value in the rightmost file.
The _contents of arrays are not merged_ as there is no sensible and consistent strategy for doing so.

For example:

`exampleapp -c a.json,b.json`

`a.json`

```json
{
  "a": [1,2,3],
  "b": ["a", "b", "c"]
}
```

`b.json`

```json
{
  "a": [4,5,6]
}
```

results in an effective configuration of:

```json
{
  "a": [1,2,3],
  "b": ["a", "b", "c"]
}
```

---
**Next**: [Logging](log-index.md)

**Prev**: [Configuration type handling](cfg-types.md)
