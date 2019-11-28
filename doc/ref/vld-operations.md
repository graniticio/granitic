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


**Next**: [Custom operations](vld-custom.md)

**Prev**: [Creating and enabling rules](vld-enable-rules.md)

