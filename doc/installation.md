# Installing Granitic

## Requirements

 * A UNIX type OS (MacOS, Linux, UNIX, BSD etc) - Granitic has not been tested against Windows.
 * Go 1.6 or later
 * Git
 
 It is highly recommended that you have installed Go according to the standard Go installation instructions 
 (https://golang.org/doc/install) and have set your GOPATH environment variable correctly.
 
## Installing the current version

Open a terminal and run:

<pre>
go get github.com/graniticio/granitic
</pre>
 
## Build the support tools

Open a terminal and run:

<pre>
go install github.com/graniticio/granitic/cmd/grnc-bind
go install github.com/graniticio/granitic/cmd/grnc-ctl
go install github.com/graniticio/granitic/cmd/grnc-project
</pre>
 
## Set the GRANITIC_HOME environment variable

You will need to set the following environment variable in your .bash_profile file (or whichever file your shell uses to
set user environment variables)

GRANITIC_HOME=$GOPATH/src/github.com/graniticio/granitic

You should also make sure that your PATH variable includes

$GOPATH/bin
 
## Testing your installation

The following commands make use of all the environment variables and some of the support tools used when developing 
Granitic applications.
 
Open a terminal and run:

<pre>
cd $GOPATH/src
grnc-project install-test
cd install-test
grnc-bind
go build ./install-test.go
./install-test
</pre>

If your installation has been successful, you'll see command line output similar to:

<pre>
04/Jan/2017:11:14:19 Z INFO  grncInit Starting components
04/Jan/2017:11:14:19 Z INFO  grncInit Ready (startup time 970.09µs)
</pre>

You can stop the program with CTRL+C and can safely delete the $GOPATH/src/install-test folder.

## Next steps

For more information on developing Granitic applications, please [work through the tutorials](https://github.com/graniticio/granitic/tree/master/doc/tutorial)