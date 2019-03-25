# The component lifecycle

Granitic is responsible for creating the Go structs that you tell it manage as components. By default, this means that
Granitic will create a new instance of the struct (using `new()`). The instance will remain in memory until your application 
shuts down.

If you want to take advantage of it, Granitic supports a more sophisticated lifecycle management pattern that your
components can opt-in to by implementing one or more of the interfaces provided in 
[Granitic's IoC package](https://godoc.org/github.com/graniticio/granitic/ioc)


## Start

If your component implements [ioc.Startable](https://godoc.org/github.com/graniticio/granitic/ioc#Startable), it's 
`StartComponent()` method will be called once Granitic has finished instantiating and configuring all
components.

This method is a good place to put any initialisation code. Any errors returned by a `StartComponent()` method will
prevent your application from starting.

## Allow access

Components that allow inbound communication (via web services, queues or some other) are encouraged to implement
[ioc.Accessible](https://godoc.org/github.com/graniticio/granitic/ioc#Accessible). The `AllowAccess()` method
on components implementing `ioc.Accessible` is called _after_ all components have their `StartComponent()` methods
invoked.

This allows you to make sure that all components in your application have started correctly and are ready before opening
ports into the application. This technique is employed by Granitic's [HTTPServer facility](fac-http-server.md)

Any errors returned by an `AllowAccess()` method will prevent your application from starting.


### Blocking the allow access phase

If your component implements [ioc.AccessibilityBlocker](https://godoc.org/github.com/graniticio/granitic/ioc#AccessibilityBlocker),
it cam prevent Granitic moving from the 'start' phase to the 'allow access' phase for _all_ components.

This can be useful if your component has spawned a goroutine to perform some sort of initialisation in it's `StartComponent()`
method. For example, the component might be registering your application with a service discovery manager or an external
logger.

If your component's `BlockAccess()` method returns true, Granitic will not progress to the `allow access` phase. The 
number of times this method will be called and the interval between checks in controlled through the ['system' facility](adm-system.md).

Id `BlockAccess()` returns an error, your application will not start.


## Suspend and resume

Components that want to be able to stop or interrupt their behaviour in reaction to the entire application being 
'suspended' should implement [ioc.Suspendable](https://godoc.org/github.com/graniticio/granitic/ioc#Suspendable).

The `Suspend()` method will be called when the application receives a suspend command through [runtime control](rtc-index.md) 
and `Resume()` will be called when a resume command is received. 

Errors returned by these methods will be logged, but do not cause the application to exit. 

## Stopping

Components that need to perform cleanup, finish work or safely dispose of resources should implement
[ioc.Stoppable](https://godoc.org/github.com/graniticio/granitic/ioc#Stoppable). This requires your component to implement
three methods.

### PrepareToStop()

`PrepareToStop()` is a warning that a stop signal is imminent. This is your component's opportunity to prevent any new
work being started (by closing ports, interrupting jobs etc.). This method should return immediately as `ReadyToStop()`
is your component's opportunity to delay shutdown.

### ReadyToStop()

`ReadyToStop()` allows your component to say whether it is now ready for the application to shut down. Any errors
returned will be logged but will not cause the application to exit.

The number of times an individual component can say that it is not ready before the application shuts down anyway are
controlled though the ['system' facility](adm-system.md).


### Stop()

This is the instruction from Granitic to stop all work immediately. It is your component's last chance to try and 
cleanly stop work or free up resources.

## Component state

If your application wants to keep track of it's current lifecycle state it can use the pre-defined 
[ioc.ComponentState enumeration](https://godoc.org/github.com/graniticio/granitic/ioc#ComponentState). This is
purely a convenience for your component the values are not read or set by the framework.

---
**Next**: [Decorators](ioc-decorators.md)

**Prev**: [Component templates](ioc-templates.md)
