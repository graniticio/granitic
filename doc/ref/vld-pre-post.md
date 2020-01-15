# Pre-validation
[Reference](README.md) | [Automatic Validation](vld-index.md)

---

It is sometimes necessary to modify data after it has been parsed into a [target object](ws-capture.md) but before validation
takes place. Your [logic component](ws-logic.md) should implement the 
[ws.WsPreValidateManipulator](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsPreValidateManipulator) interface.

Granitic will call the `PreValidate()` method and your implementation can access the target object via the 

**Next**: [Relational databases](db-index.md)

**Prev**: [Custom operations](vld-custom.md)