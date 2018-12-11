# Introduction to Granitic

The purpose of Granitic is to allow developers who want to build web services in Go to do so by using a single library.
Granitic is a framework (a library of code and patterns) but it is also a lightweight application server. By using Grantic 
your Go application becomes a self-contained service which can accept HTTP requests, run scheduled activities or perform
any other task that is appropriate for the service tier of your architecture.

## Who should use Granitic?

There are many service and micro-service frameworks available for Go, each with its own intended audience and use cases. 
Granitic is aimed at:

 * Developers who are new to Go and are moving from a framework like Spring or Rails in another programming language.
 * Developers who want a single framework to provide 'enough' features to build production ready web services without needing to
 research and import multiple libraries.
 * DevOps practitioners who require services to be container and FaaS friendly and to exist within a configuration-management controlled environment.
 
  
## Principles

Granitic was developed according to a number of design principles that are worth understanding in order to give context to the way Granitic works
and help you decide if it is right for your project.

###  Inversion of Control for lifecycle management

For building services where a non-trivial number of objects are created and where the concept of 'global' or 'singleton' objects are important,
Inversion of Control (IoC) remains a key design pattern. IoC is commonly implemented using a container, where object life-cycles and configuration
are defined declaratively in configuration files or as code annotations.

This model is common in languages where types can be loaded at runtime (like Java, C# etc) but less common in statically linked languages like Go.

The heart of Granitic is a fully featured IoC container, which allows user components to be defined in configuration files and converted into code using
a 'binding' step in the build process.

###  Externalised and layered configuration

Typically services are created in a development environment then transition through a number of test and pre-production environments before eventually
being deployed to production. In each environment, different behaviours will required (e.g. use of a mocked dependency in development) and different external
resources will be accessed (e.g. different database server hostname).

Granitic recognises and supports this workflow in two ways:

 * Configuration files are separate from the executable service (e.g. they're not 'burnt in' at compile time)
 * Configuration files can be layered (e.g. the configuration common to all environments can be defined in one file, the changes/overrides required
 for a particular environment can be defined in another)
 
 Additionally, Granitic does not require that configuration files are hosted in the same compute resource (FaaS, container, VM) as the executable. They
 can be loaded from a remote source at startup time.

### No downstream dependencies

Package management is a complicated problem in any programming language and it is our belief that it is developers who should
choose which third-party libraries at which versions should be included in their application, not framework developers. As such Granitic has no downstream dependencies at all, even
test frameworks.

### Easily replaceable core features

Because Granitic is intended to be self-contained, it has implementations of many capabilities (logging, scheduling, validation) that
are the subjects of other large and fully featured Go libraries. Granitic's principle is that its implementation of these capabilities
(facilities, in Granitic) should be mature and fully featured enough to support the vast majority of use-cases, but if you want to swap out
Granitic's implementation with your preferred third-party implementation, you will be supported in doing so.

### Fast start-up

Cold start time is an increasing concern to application developers due to the increasing use of containers and FaaS as a deployment mechanism. Start up time
is a key performance metric for Granitic and is actively monitored and designed for. Simple Granitic applications have sub-millisecond startup times on modern
hardware.  