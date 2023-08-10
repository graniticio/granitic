# Installing Granitic

## Requirements

 * Go 1.17 or later
 * Git


## GOPATH

Granitic relies on your [GOPATH environment variable](https://go.dev/doc/gopath_code#GOPATH) being set correctly. This should be set to the directory where Go 
downloads packages and installs binaries and is normally set to `$HOME/go`.

## Downloading Granitic and setting GRANITIC_HOME

When building Granitic applciations, you will need to have the Granitic source tree checked out to your local filesytem
and that location defined in an environment variable `GRANITIC_HOME`. Open a terminal and run:

<pre>
git clone https://github.com/graniticio/granitic ~/granitic
export GRANITIC_HOME=~/granitic
</pre>

Make sure to also define `GRANITIC_HOME` in your shell's environment scripts (`.bash_profile`, `.zshenv` etc)

You can now install Granitic's command line tools by running:

<pre>
cd $GRANITIC_HOME
cmd/install-tools.sh
</pre>


Which will install the tools `grnc-bind` and `grnc-project` in `$GOPATH/bin` or in `$HOME/go/bin`, Make sure that 
`$GOPATH/bin` is included in your `PATH` environment variable. 


### Building and running a test application

Open a terminal and run:

<pre>
cd /tmp
grnc-project install-test
cd install-test
go mod tidy
grnc-bind && go build
./install-test
</pre>

If your installation has been successful, you'll see command line output similar to:

<pre>
02/Aug/2023:09:11:19 Z INFO  [grncInit] Granitic v2.2.2
02/Aug/2023:09:11:19 Z INFO  [grncInit] Starting components
02/Aug/2023:09:11:19 Z INFO  [grncInit] Ready (startup time 2.090459ms)
</pre>

You can stop the program with CTRL+C and can safely delete the /tmp/install-test folder.

## Troubleshooting

If you have problems installing or running the support tools. Make sure that:
  
  * Your `GOPATH` environment variable [is set correctly](https://github.com/golang/go/wiki/GOPATH) and contains a folder called `bin` _or_ the `GOBIN` environment
  variable is set.
  * Your `PATH` variable includes `$GOPATH/bin` or `$GOBIN`

## Next steps

For more information on developing Granitic applications, please [work through the tutorials](https://github.com/graniticio/granitic/v2/tree/master/doc/tutorial)