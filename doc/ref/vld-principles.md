# Automatic Validation Principles
[Reference](README.md) | [Automatic Validation](vld-index.md)

## Minimising boilerplate code

Implementing validation of user data submitted to web services generally results in lot of boilerplate code,
especially in a statically typed language like Go that doesn't support concepts like annotations.

A primary goal of Granitic's validation framework is to remove as much of this boilerplate code as possible.

## Validation as rules

Rather than implementing validation as imperative code, Granitic allows the various checks (_operations_, in 
Granitic terminology) performed on a field submitted by a user to be expressed as a terse rule.

These rules are defined in Granitic's standard [configuration files](cfg-files.md) and as such can be modified 
without recompiling your application and can have their definition changed between envrionments, if required.

## Integration with processing pipeline

Automatic validation is fully integrated with the [web service processing pipeline](ws-pipeline.md). Its goal
is to deliver a validated [data capture target](ws-capture.md) to your application logic or automatically
return an [error response](ws-error.md) if validation failed. 
  
## Custom operations

Granitic includes a number of common operations for checking various data-types, but also provides a mechanism
for you to build custom operations for additional checks (e.g. making sure an email address
is not already being used by another user)

---
**Next**: [Creating and enabling rules](vld-enable-rules.md)

**Prev**: [Automatic validation index](vld-index.md)