# JSON and YAML

Granitic uses JSON for configuration and component definition files. This choice is partly driven by features, but is mostly driven by the design principle that Granitic will not introduce any dependencies on your application that you do not have control over. Go has built-in JSON support, so JSON can be used without violating this principle.

## YAML

YAML is considered by many to be superior to JSON for configuration files that need to be viewed and edited by people, principly due the reduced need for structural markers like brackets and quotes as well as more flexibility around how structures like lists and maps are represented.

Unfortunately Go does not natively include a YAML parser, so Granitic has as sister project [granitic-yaml](https://github.com/graniticio/granitic-yaml) which enables full support for YAML configuration and component definition files in Granitic at the expense of adding a dependency on the external [yaml](https://gopkg.in/yaml.v2) package.

---
**Next**: [Building and running your application](gpr-build.md)

**Prev**: [The component container](ioc-index.md)