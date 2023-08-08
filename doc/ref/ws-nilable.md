# Nilable types

[Reference](README.md) | [Web Services](ws-index.md)

---

A drawback of Go's system of [zero values](https://tour.golang.org/basics/12) is that it introduces ambiguity:
is the value of a field on your target object `false` because the web service client set the value to false, or because
the value wasn't supplied at all and Go defaulted the value to false?

Granitic gets around this problem by providing a set of ['nilable' types](https://godoc.org/github.com/graniticio/granitic/v2/types)
for [bool](https://godoc.org/github.com/graniticio/granitic/v2/types#NilableBool), [string](https://godoc.org/github.com/graniticio/granitic/v2/types#NilableString),
[int64](https://godoc.org/github.com/graniticio/granitic/v2/types#NilableInt64) and [float64](https://godoc.org/github.com/graniticio/granitic/v2/types#NilableFloat64).

These types all implement the interface [types.Nilable](https://godoc.org/github.com/graniticio/granitic/v2/types#Nilable)
which provides a method `IsSet() bool` which can tell you if the value was provided by the client (true) or not (false).

Granitic has deep support for these types and the framework can use them interchangeably in the whole [web request processing cycle](ws-pipeline.md), 
including body parsing, path binding, query binding and JSON/XML marshalling. Other parts of Granitic including [rule based validation](vld-index.md)
and [query management](fac-query.md) also support these types.

## Using nilable types in your target object

Given a target object:

```go
type MyTarget struct {
  IntField    int
  BoolField   bool
  StringField string
  FloatField  float
}
```

Rewrite this as:

```go
type MyTarget struct {
  IntField    *types.NilableInt64
  BoolField   *types.NilableBool
  StringField *types.NilableString
  FloatField  *types.NilableFloat64
}
```

Note that the fields must be declared as _pointers_ to the nilable types.


## Using values

Each of the nilable types provides a method to recover the value it contains. The method is named according to the type
it returns - e.g. [types.NilableBool.Bool()](https://godoc.org/github.com/graniticio/granitic/v2/types#NilableBool.Bool)



---
**Next**: [Validating data](ws-validate.md)

**Prev**: [Capturing data](ws-capture.md)