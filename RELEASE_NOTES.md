# Granitic 3.0

Granitic 3.0 makes changes that may not be compatible with your version 2.x projects. See the
'Breaking Changes' section at the end of this document.

## Go 1.19

Granitic 3.0 requires Go 1.19 or later.

## YAML configuration as default

Granitic 3.0 now uses YAML files as the default for both configuration files and component definition files.
JSON files are still supported in user applications and third party libaries and your project may include
a mix of JSON and YAML files as required.

## Change to 'no dependencies' policy

In order to modernise Granitic, we have had to relax our policy on importing third party libraries into the 
core Granitic source code. Our policy is now to minimise the use of third party libraries and only use those
libraries that are trusted by major Go projects such as Kubernetes.

Version 3 now imports [YAML](https://pkg.go.dev/gopkg.in/yaml.v3) and [gocheck](https://labix.org/gocheck)

## External facilities

Version 3 now allows you to include component definitions and default configuration for external libraries
in your application. This allows teams building services that rely on shared or third libraries to avoid having to
manually import component definition files into every service that uses that library.

## Change of focus

Since Granitic was originally created, the way micro services are built, deployed and run has fundamentally
changed. The original purpose of Granitic was to help Java developers to migrate large, XML-centric Spring
applications into Go.

While that is still a supported use case, the most common use of Granitic is to build small micro services running
in containers or as a Function-as-a-Service wrapper like AWS Lambda.

For that reason, several built-in facilities have been moved into external facilities:

  * Runtime Control (`RuntimeCtl`)
  * Task Scheduler (`TaskScheduler`)
  * XML Web Services (`XMLWs`)


You will need to change your application if you want to keep using them (information on how to do this is 
in the Breaking Changes) section.

## Breaking Changes

TBC