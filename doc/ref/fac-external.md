# External Facilities
[Reference](README.md) | [Facilities](fac-index.md)

---

External facilities allow you to define your own modules of code, component definitions and default 
configuration that behave in a similar way to Granitic's built-in facilities. 

## How this works

1. Running `grnc-bind` with the optional `-x` flag instructs Granitic to inspect the modules imported via 
your application's `go.mod` file.
2. For each module in `go.mod`, Granitic locates the source files for that module and tries to find a manifest file (see below)
3. The manifest file will declare one or more external facilities with component definitions and default configuration.
4. The external facility components will be included in your application in the same way as the components you have defined
yourself (e.g. they will be converted into Go source code in your [bindings file](gpr-build.md))
5. The default configuration for the external facility will be serialised and included in your application's executable.
6. Each external facility can optionally declare a 'builder' component. This component will be invoked before
the IoC container is started and allows for components to be created and modified programmatically. 
7. Each external facility can declare whether it is in an enabled or disabled state by default and a simple 
configuration path (see below) is defined to allow applications to override this default state with configuration.

## Converting your  Go module to an external facility

Any Go module can be converted into an external facility by creating a folder called `facility` in the root of
the module's source tree and adding to it a file called `manifest.json`. The folder can also optionally contain two 
sub-folders `comp-def` and `config` for storing component definitions and configuration files respectively

```shell
  /facility
    manifest.json
    /config
      *.json
    /comp-def
      *.json
```

### Using YAML instead of JSON

If you are using [grnc-yaml](http://github.com/graniticio/granitic-yaml), you can create your external facility's 
manifest, component definition and configuration files as YAML instead of JSON. Your application can 
import external facilities that use a mixture of JSON and YAML files.

### Alternative to facility folder

If your module already has a folder at the module root level called `facility`, you can use the alternative
name `grnc-facility`

## Manifest files

A typical manifest file may look like this:

```json
{
  "Namespace": "MyCompany",
  "Facilities": {
    "MyFirstFacility": {
      "Disabled": true,
      "Depends": ["HTTPServer", "JSONWs", "OtherVendor.Tracing"],
      "Builder": "mypackage/sub"
    },
    "MySecondFacility": {}
  }
}
```

### Namespace

To avoid clashes between the names of external facilities from multiple vendors, each manifest file
must declare a 'namespace'. This is typically the name of your organisation as a valid JSON identifier.

### Facilities

This is a map containing each of the external facilities defined in the manifest. The key for each map entry
(`MyFirstFacility` and `MySeondFacility` in the above example) is the name of the facility.

#### Canonical names

The canonical name of each facility is `Namespace.FacilityName` (e.g. `MyCompany.MyFirstFacility`) this is
how the facility will be referenced in config files, dependencies etc.

### Disabled

An external facility can be disabled or enabled by default. This field is called `Disabled` rather than `Enabled`
so that if the field is not defined, the external facility is enabled (e.g. the value of `Disabled` will implicitly be false).

#### Accessing the value of this field

The value of this field will be _inverted_  and made available to your application's configuration as:

`ExternalFacilities.Namespace.FacilityName`

So if `ExternalFacilities.MyCompany.MySecondFacility` is set to `true`, your application knows this facility is enabled. 
Your application could choose to disable the facility by including this in its configuration:

```json
{
  "ExternalFacilities": {
    "MyCompany": {
      "MySecondFacility": false
    }
  }
}
```

#### The effect of enabling or disabling and external facility

In practical terms, 'disabling' an external facility triggers the following behaviour:

  1. If the external facility declares a 'builder' (see below), the builder will not be invoked.
  2. The value of `ExternalFacilities.Namespace.FacilityName` will be `false`

If your facility is disabled by default, any components declared in the facility's component definition files will still
be instantiated and any configuration will still be included in your application's executable.


---
**Next**: [Runtime Control](rtc-index.md)

**Prev**: [Service Errors](fac-service-errors.md)