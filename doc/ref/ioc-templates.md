# Component templates
Back to: [Reference](README.md) | [Component Container](ioc-index.md)

---

In larger applications, many components can share the same type and some of the same dependencies. The
most common example of this is web service handlers, which are implicitly all the same type and will often share the same
components for [identity and access management](ws-iam.md). To avoid having to redeclare this, you can use component templates.

A component template can define a type, configuration and dependencies like a component, but it is not instantiated by 
Granitic. Instead it is referred to by other components.

## Declaring a component template

Component templates are declared in the `templates` JSON object/map at the root of your component definition file:

```json
{
  "templates": {
    "handler": {
      "type": "handler.WsHandler"
    }
  }  
}
```

Instead of specifying a `type` your component can now use the `compTemplate` field (or `ct` if `compTemplate` is too
verbose) to inherit properties from a template:

```json
{
  "components": {
    "myHandler": {
      "ct": "handler"
    }
  }
}
```


## Child templates

Templates can inherit from a single parent template using the `compTemplate` or `ct` instead of a `type`. In this way you can set up chains of templates.

For example:

```json
{
  "templates": {
    "handler": {
      "type": "handler.WsHandler"
    },
    
    "secureHandler": {
      "ct": "handler",
      "AccessChecker": "+myAccessChecker",
      "RequireAuthentication": true
    },
    
    "secureGetHandler": {
      "ct": "secureHandler",
      "HTTPMethod": "GET"
    }
  },
  
  "components": {
    "myHandler": {
      "ct": "secureGetHandler"
    }
  }  
}
```

---
**Next**: [Lifecycle management](ioc-lifecycle.md)

**Prev**: [Component definition files](ioc-definition-files.md)