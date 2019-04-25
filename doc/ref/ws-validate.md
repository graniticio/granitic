# Validating data

Back to: [Reference](README.md) | [Web Services](ws-index.md)

---

After data is [captured](ws-capture.md) from a client's HTTP request into a target object, 
[handler.WsHandler](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsHandler) next performs validation
of the captured data before it is passed onto your [logic component](ws-logic.md).

Each step of the validation process is optional and by default no validation is peformed. The logical steps are:

 * Pre-validation manipulation
 * Automatic rule-based validation
 * Manual code-based validation

## Pre-validation

Under some circumstances your code might want to modify the [captured data](ws-capture.md) before it is validated.
Possible reasons to do this include:
  
  * Normalisation of data
  * Cleanup of messy data if your API is quite permissive about what is submitted
  
If your code requires this step, you must create a component that implements 
[handler.WsPreValidateManipulator](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsPreValidateManipulator)
and set a reference to that component on the `PreValidateManipulator` field of your handler.

Granitic will call the `PreValidate` method you define. If your code encounters problems and decides that
validation should not proceed, it should return `false` - and ideally add [service errors](ws-error.md) to
the [ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors) object that is passed in.


## Auto-validation

Data captured into a target object can be automatically validated by Granitic using rules you define in configuration.
This is a major feature of Granitic and is documented in its [own section of the reference manual](vld-index.md).

## Manual validation

If you do not want to use [automatic validation](vld-index.md) or if you have complicated validation scenarios
that automatic validation does not support, you can perform manual validation where you write code for the
checks you need to make.

To do this, your [logic component](ws-logic.md) must implement 
[handler.WsRequestValidator](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestValidator)

Granitic will call the `Validate` method you define. Any validation errors encountered should be recorded
by adding [service errors](ws-error.md) to the 
[ws.ServiceErrors](https://godoc.org/github.com/graniticio/granitic/ws#ServiceErrors) object that is passed in.
          
### Automatic and manual validation

There are some circumstances under which you will want to define some validation as rules to be automatically
applied and some additional validation to be applied manually. This is supported, but you will need to 
be aware of the `DeferAutoErrors bool` field on your handler.

The default setting is `false` which means that if automatic validation fails, manual validation will not be 
applied and your [handler.WsRequestValidator.Validate()](https://godoc.org/github.com/graniticio/granitic/ws/handler#WsRequestValidator)
method will not be called.

Setting the field to `true` means that your manual validation will always be called after automatic validation (pass or fail),
but your manual code needs to be aware that the data may be in an inconsistent state (e.g. some fields might be missing).


---
**Next**: [Application logic](ws-logic.md)

**Prev**: [Nilable types](ws-nilable.md)