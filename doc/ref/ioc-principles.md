## Component Container Principles

## Single named instance

A component is an instance of some Go struct that has been assigned a name by which other components can refer to it. 
That name must be unique - two components cannot have the same name.

## Application and framework components

Your application's components (those that you define in component definition files) run alongside the components
that the Granitic framework runs to provide [facilities](fac-index.md) to your application. By convention the names
of the Granitic components start with the prefix `grnc`

## Granitic manages dependency injection

Your components should not create instances of the other components they rely on. Instead, your component definition
files should explicitly define the dependencies between components and Granitic will make sure the components you need
are injected at runtime.

## Lifecycle events

You component can optionally opt-in to receiving notifications about changes to the state of the container (such as starting,
getting ready to shutdown etc). Your components do this by [implementing specific interfaces](ioc-lifecycle.md)

## Goroutine safe

As there is only one instance of each component shared among all users accessing your application, it is vital that 
your code is goroutine safe. If you components store state (say in member variables) it is vital that modifications 
to that state are performed in a safe manner.

## Runtime control

The component container be interacted with at runtime via the [built-in runtime control commands](rtc-built-in.md)

---
**Next**: [Component definition files](ioc-definition-files.md)

**Prev**: [Component container index](ioc-index.md)
