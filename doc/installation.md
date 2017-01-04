# Installing Granitic

## Requirements

 * A UNIX type OS (MacOS, Linux, UNIX, BSD etc) - Granitic has not been tested against Windows.
 * Go 1.6 or later
 * Git
 
 It is highly recommended that you have installed Go according to the standard Go installation instructions 
 (https://golang.org/doc/install) and have set your GOPATH environment variable correctly.
 
## Installing the current version

 1. Open a terminal
 2. go get github.com/graniticio/granitic
 
## Build the support tools

 1. Open a terminal
 2. go install github.com/graniticio/granitic/cmd/grnc-bind
 3. go install github.com/graniticio/granitic/cmd/grnc-ctl
 4. go install github.com/graniticio/granitic/cmd/grnc-project
 
## Set the GRANITIC_HOME environment variable

You will need to set the following environment variable in your .bash_profile file (or whichever file your shell uses to
set user environment variables)

GRANITIC_HOME=$GOPATH/src/github.com/graniticio/granitic

You should also make sure that your PATH variable includes

$GOPATH/bin
 
## Testing your installation

The following commands make use of all the environment variables and some of the support tools used when developing 
Granitic applications.
 
 1. Open a terminal
 2. cd $GOPATH/src
 3. grnc-project install-test
 4. cd install-test
 5. grnc-bind
 6. go build ./install-test.go
 7. ./install-test
 
If your installation has been succesful, you'll see command line output similar to:

04/Jan/2017:11:14:19 Z INFO  grncInit Starting components
04/Jan/2017:11:14:19 Z INFO  grncInit Ready (startup time 970.09µs)

You can stop the program with CTRL+C and can safely delete the $GOPATH/src/install-test folder.

## Next steps

For more information on developer Granitic applications, please work through the tutorials at:

https://github.com/graniticio/granitic/doc/tutorial