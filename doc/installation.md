# Installing Granitic

## Requirements

 * Go 1.8 or later
 * Git
 
 It is highly recommended that you have installed Go according to the [standard Go installation instructions](https://golang.org/doc/install) and have set your GOPATH environment variable correctly.
 
#### Note for Windows users
 
 The below instructions were tested on Windows 10 having followed the [Go installation instructions for Windows](https://golang.org/doc/install?download=go1.7.5.windows-amd64.msi)
 
 You must [set the GOPATH environment variable](https://golang.org/doc/code.html#GOPATH) and have the Git command line tools installed and configured to work with Command Prompt
 
## Installing and testing

### Installing the current version of Granitic

Open a terminal and run:

<pre>
go get github.com/graniticio/granitic
</pre>
 
### Build the support tools

Open a terminal and run:

<pre>
go install github.com/graniticio/granitic/cmd/grnc-bind
go install github.com/graniticio/granitic/cmd/grnc-ctl
go install github.com/graniticio/granitic/cmd/grnc-project
</pre>
 
### Set the GRANITIC_HOME environment variable

#### UNIX-type operating systems

You will need to set the following environment variable in your .bash_profile file (or whichever file your shell uses to
set user environment variables)

<pre>GRANITIC_HOME=$GOPATH/src/github.com/graniticio/granitic</pre>

You should also make sure that your PATH variable includes _$GOPATH/bin_

### Windows

You need to create an environment variable named GRANITIC_HOME and set it to:

<pre>
%GOPATH%\src\github.com\graniticio\granitic
</pre>

You should also make sure that your PATH environment variable includes _$GOPATH\bin_
 
### Testing your installation

The following commands make use of all the environment variables and some of the support tools used when developing 
Granitic applications.
 
#### UNIX-type operating systems 
 
Open a terminal and run:

<pre>
cd $GOPATH/src
grnc-project install-test
cd install-test
grnc-bind
go build ./service.go
./service
</pre>

### Windows

<pre>
cd %GOPATH%\src
grnc-project install-test
cd install-test
grnc-bind
go build service.go
service
</pre>


If your installation has been successful, you'll see command line output similar to:

<pre>
04/Jan/2017:11:14:19 Z INFO  grncInit Starting components
04/Jan/2017:11:14:19 Z INFO  grncInit Ready (startup time 970.09µs)
</pre>

You can stop the program with CTRL+C and can safely delete the $GOPATH/src/install-test folder.

## Next steps

For more information on developing Granitic applications, please [work through the tutorials](https://github.com/graniticio/granitic/tree/master/doc/tutorial)