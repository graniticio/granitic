# Operations Reference
 [Reference](README.md) | [Automatic Validation](vld-index.md)
 
This page describes the operations available as part of a validation rule. 

## Common operations

These operations are available for more than one [field type](vld-enable-rules.md)

---

### REQ (Required)

`REQ[:ERROR_CODE]`

#### Available for 

All types

#### Parameters

No parameters

#### Usage

The `REQ` operation indicates that the field being checked must have been 'set'. The definition
of set changes according to the underlying Go data type:

 * For basic Go types (string, intX, floatX etc) `REQ` will always pass
 * For [nilable types](ws-nilable.md), 'object' and 'slice' types, `REQ` will fail if the caller did not explicitly provide 
 a value for the field.

---

### MEX (Mutally Exclusive)

`MEX:field1,field2...fieldN[:ERROR_CODE]`

#### Available for 

All types

#### Parameters

`MEX` requires a comma separated list of one or more field names.

#### Usage

The `MEX` operation allows you to declare that a field is not allowed to be set if _any_ of the other
fields specified have been set. 

For example `MEX:WeightLbs,WeightStones` would fail if _either_ `WeightLbs` or `WeightStones` is set. The logic
for whether a field is considered to be 'set' is the same as the logic described above in the `REQ` operation.

---

### BREAK 

`BREAK`

#### Available for 
 * INT
 * FLOAT
 * STR
 
#### Parameters

None

#### Usage

`BREAK` stops execution of the current rule if any check operations have already failed for that rule. `BREAK` is used
to prevent expensive or irrelevant checks being made on a field that is already known to be invalid in some way.

---

### STOPALL

`STOPALL`

#### Available for 

All types

#### Parameters

None

#### Usage

The presence of `STOPALL` in a rule indicates that no _other rules_ should be run if that rule fails. 

---

### IN

`IN:value1,value2...valueN[:ERROR_CODE]`


#### Available for 
 * INT
 * FLOAT
 * STR
 
#### Parameters

`IN` requires a comma separated list of one or more values.

#### Usage

`IN` fails if the value of the field being checked is not one of the values specified in the list of allowed values.

For `FLOAT` rules, the values in the allowed list are parsed as `float64` and `int64` for `INT` rules.

Please note that there is currently no support for escaping the value list separator (comma) for `STR` rules.

---

### LEN (Length)

`IN:[min]-[max][:ERROR_CODE]`

#### Available for
  * STR
  * SLICE
  
#### Parameters

`LEN` requires a minimum length, a maximum length or both

#### Usage

`LEN` checks either the number of characters (not bytes) in a string or the elements in a slice/array. You can specify a 
minimum and maximum length (`LEN:2-10`), just a maximum (`LEN:-10`) or just a minimum (`LEN:2-`)

---

### EXT (External component)

`EXT:componentName[:ERROR_CODE]`

#### Available for
  * STR
  * INT
  * FLOAT
  
#### Parameters

`EXT` requires the component name of a [component registered in the Granitic IoC container](ioc-principles.md)
that can determine whether or not the field under validation is allowed.

#### Usage

`EXT` allows complex checks that cannot be modelled using Granitic's rules and operations or require access to an external
data source. For example, you might want to check if an email address is already being used by another user before
allowing another user to claim it.

The component that is used to perform the check must implement one of [validate.ExternalStringValidator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalStringValidator),
[validate.ExternalFloat64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalFloat64Validator) or
[validate.ExternalInt64Validator](https://godoc.org/github.com/graniticio/granitic/validate#ExternalInt64Validator) according to 
the type of the field under consideration.

---

### RANGE

`RANGE:[min]|[max][:ERROR_CODE]`

#### Available for
  * INT
  * FLOAT
  
#### Parameters

`RANGE` requires a minimum allowed value or a maximum allowed value or both.

#### Usage

`RANGE` allows to check whether a numeric value is within a pre-defined range. As the minimum and maximum parameters are
both optional, this operation can also be used to implement minimum (with no maximum) or maximum (with no minimum) value checks.

*Examples:*

`RANGE:1|10` value must be between 1 and 10 inclusive

`RANGE:-2|` value must be a minimum of -2, with no upper limit

`RANGE:|0.98` value must be a maximum of 0.98 with no lower limit

--- 

## BOOL operations

The following operations are only available for checks on `BOOL` fields.

### IS

`IS:boolValue[:ERROR_CODE]`

#### Parameters

`IS` takes either `true` or `false` as a parameter

#### Usage

`IS` is used to require a boolean value to be set to the same value (true or false) as specified in the check.

---

## STRING operations

The following operations are only available for checks on `STRING` fields.

### TRIM

`TRIM`

#### Usage

`TRIM` temporarily removes leading and trailing whitespace (using [strings.TrimSpace()](https://golang.org/pkg/strings/#TrimSpace))
for the purposes of all validation checks. If the value passes all checks, it is the original untrimmed version of the value which
is passed to your [application logic](ws-logic.md).

---

### HARDTRIM

`HARDTRIM`

#### Usage

`HARDTRIM` behaves in a similar way to `TRIM`, except that the value is permanently changed to its trimmed value 
(e.g your [application logic](ws-logic.md) will receive the trimmed version of the value.

---


#### REG (regular expression)

`REG:expression[:ERROR_CODE]`

#### Parameters

`REG` requires a mandatory [regular expression pattern](https://golang.org/pkg/regexp/). The validity of the expression
is checked during Granitic startup with [regexp.MustCompile()](https://golang.org/pkg/regexp/#MustCompile). If this 
fails your application will shutdown with an error.
 
#### Usage

The string value being checked will be matched against the compiled pattern using [regexp.MatchString()](https://golang.org/pkg/regexp/#MatchString)
and the check will fail if there is no match.

---

## SLICE operations

The following operations are only available for checks on `SLICE` fields.

**Next**: [Custom operations](vld-custom.md)

**Prev**: [Creating and enabling rules](vld-enable-rules.md)

