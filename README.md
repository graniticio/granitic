# Granitic

Granitic is a framework and platform for building and deploying web services and micro services in Go.

It features:

* A web service aware HTTP server with support for load-management and Identity Access Management (IAM) integration.
* A fully-featured Inversion of Control (IoC) container.
* A flexible and customisable request processing pipeline including:
    * Full support for JSON and plain XML web services.
    * Automatic binding of request bodies, query parameters and path parameters.
    * Declarative, rule driven validation.
    * A comprehensive error management system including full templating of all system and application error messages and HTTP response code mapping.
* Component based error logging.
* Query management for data sources.
* RDMBS integration with an interface designed to promote more readable code.


Additionally, Granitic is designed to be 'DevOps friendly' and offers:

* Fully externalised configuration, with support for configuration files stored locally or served over HTTP.
* Low memory footprint and fast startup times (compared to JVM/CLR equivalents).
* Runtime control of deployed applications (including suspension/resumption).
* Runtime control of log levels (e.g. temporarily enable debugging without restarts).

