# Shared rules
[Reference](README.md) | [Automatic Validation](vld-index.md)

---

Rules are typically defined in the context of a single endpoint in the auto validator component that is 
[injected into a handler](vld-enable-rules.md). There are scenarios in which is is useful to define rules
in way they can be shared between multiple handlers:

  * A common rule that is needed by multiple endpoints (email address validation, for example)
  * A rule that is used by another operation (the [ELEM](vld-operations.md) operation for `SLICE` validation)
  
There are a number of steps to enabling shared rules:

## Shared rule manager

Create a component of type `validate.UnparsedRuleManager` and provide a config path to your rule definitions:

```json
"sharedRuleManager": {
  "type": "validate.UnparsedRuleManager",
  "Rules": "$sharedRules"
}
```

## Rule definitions

Define your shared rules in any configuration file. The structure is a map of string arrays like:

```json
"sharedRules": {
  "artistExistsRule": ["INT", "EXT:artistExistsChecker"],
  "personAgeRule": ["INT", "RANGE:0-120"]
}
```

The key in the map (`artistExistsRule`) is the name of the rule that can be referenced by other components. The structure
and content of the rules are defined in the same way as non-shared rules.


## Set rule manager on auto validator

The auto validator which manages validation for a particular handler must be given a reference to the shared rule manager

```json
"submitRecordHandler": {
  "type": "handler.WsHandler",
  "AutoValidator": {
    "type": "validate.RuleValidator",
    "DefaultErrorCode": "INVALID_RECORD",
    "Rules": "$createRecordRules",
    "RuleManager": "+sharedRuleManager"
  }
}
```

**Next**: [Relational databases](db-index.md)

**Prev**: [Operations reference](vld-operations.md)
