# Granitic

Granitic is a framework and lightweight application server for building and running web and micro services written in Go.

## Features

* A web service aware HTTP server with support for load-management and Identity Access Management (IAM) integration.
* A fully-featured Inversion of Control (IoC) container.
* A flexible and customisable request processing pipeline including:
    * Full support for JSON and plain XML web services.
    * Automatic binding of request bodies, query parameters and path parameters.
    * Declarative, rule driven validation.
    * A comprehensive error management system including full templating of all system and application error messages and
    HTTP response code mapping.
* Component based error logging.
* Query management for data sources.
* RDMBS integration with an interface designed to promote more readable code.
* A flexible task scheduler


Additionally, Granitic is designed to be 'DevOps friendly' and offers:

* Fully externalised configuration, with support for configuration files stored locally or served over HTTP.
* Low memory footprint and fast startup times (compared to JVM/CLR equivalents).
* Runtime control of deployed applications (including suspension/resumption).
* Runtime control of log levels (e.g. temporarily enable debugging without restarts).
* Runtime monitoring and control of scheduled activities.

## Getting started

Read and follow the [installation instructions](https://github.com/graniticio/granitic/blob/master/doc/installation.md)
then work through [the tutorials](https://github.com/graniticio/granitic/tree/master/doc/tutorial)

The GoDoc for Granitic can be [found here](https://godoc.org/github.com/graniticio/granitic)

Visit the [granitic.io](https://granitic.io) website to read the complete [reference manual](https://granitic.io/ref)

## Testing

Once you have followed the installation instructions, you can run Granitic's unit tests by running

```
cd $GOPATH/src/github.com/graniticio/granitic
go test ./...
```
## Contributing

We are always delighted to receive contributions from the community. Please read the 
[Contributing](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) guides and
then have a look at the [open issues](https://github.com/graniticio/granitic/issues)

## Get in touch

 * Twitter: @GraniticIO
 * Email: comms@granitic.io