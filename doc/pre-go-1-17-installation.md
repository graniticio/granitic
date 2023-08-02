# Installing Granitic (Go 1.11-Go 1.16)

This page explains how to install Granitic when you are required to use a version of Go earlier than 1.17

If you are able to, it is recommended that you use Go 1.17 or later and [follow these installation instructions
instead](installation.md).


## Requirements
* 
* Go 1.11 or later
* Git

It is highly recommended that you have installed Go according to the [standard Go installation instructions](https://golang.org/doc/install)
and have set your `GOPATH` environment variable correctly.

#### Note for Windows users

The below instructions were tested on Windows 10 having followed the [Go MSI installation instructions for Windows](https://golang.org/doc/install)

You must [set the GOPATH environment variable](https://golang.org/doc/code.html#GOPATH) and have the Git command
line tools installed and configured to work with Command Prompt

## Installing and testing

### Installing the current version of Granitic

Open a terminal and run:

<pre>
go get github.com/graniticio/granitic
</pre>

### Install the support tools

Open a terminal and run:

<pre>
cd $GOPATH/src/github.com/graniticio/granitic
cmd/install-tools.sh
</pre>


### Testing your installation

The following commands make use of all the environment variables and some of the support tools used when developing
Granitic applications.

#### UNIX-type operating systems

Open a terminal and run:

<pre>
cd /tmp
grnc-project install-test
cd install-test
grnc-bind
go build
./install-test
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
04/Jan/2019:11:14:19 Z INFO  grncInit Starting components
04/Jan/2019:11:14:19 Z INFO  grncInit Ready (startup time 1.749365ms)
</pre>

You can stop the program with CTRL+C and can safely delete the /tmp/install-test folder.

## Troubleshooting

If you have problems installing or running the support tools. Make sure that:

* Your `GOPATH` environment variable [is set correctly](https://github.com/golang/go/wiki/GOPATH) and contains a folder called `bin` _or_ the `GOBIN` environment
  variable is set.
* Your `PATH` variable includes `$GOPATH/bin` or `$GOBIN`

## Next steps

For more information on developing Granitic applications, please [work through the tutorials](https://github.com/graniticio/granitic/v2/tree/master/doc/tutorial)