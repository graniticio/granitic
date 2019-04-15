# Component decorators
Back to: [Reference](README.md) | [Component Container](ioc-index.md)

---

If your component implements [ioc.ComponentDecorator](https://godoc.org/github.com/graniticio/granitic/ioc#ComponentDecorator)
it will be considered a _decorator_ and will be integrated into the container startup process.

After the container has instantiated and configured all registered components, but before any [StartComponent()](ioc-lifecycle.md)
methods are called, all decorators will be given access to all other registered components.

This technique is useful in handful of situations:

  * You need to inject a component into a large number of other components (this is how [Granitic loggers are injected](log-principles.md))
  * You need to find a component of a particular type, but do not know its name ahead of time (this is how [DbProviders](db-provider.md) 
  and [instrumentation managers](ws-instrumentation.md) are discovered)

The interface requires you to implement two methods:

## OfInterest

Your decorator will have every registered component passed (one at a time) to its `OfInterest` method. The component
is represented by an [ioc.Component struct](https://godoc.org/github.com/graniticio/granitic/ioc#Component) which
provides access to the component's name and the instance of the struct the component represents.

Your decorator can decide whether the supplied component is something the decorator wants to either modify or gain
access to, usually by checking the component's type.


## DecorateComponent

If your `OfInterest` method returns `true`, the `DecorateComponent()` method is called. This is where your decorator can
modify the supplied component or capture a reference to it.

---
**Next**: [Configuration](cfg-index.md)

**Prev**: [Component lifecycle](ioc-lifecycle.md)