# Tutorial - Before you start

Make sure you have followed the Granitic [installation instructions](https://github.com/graniticio/granitic/blob/master/doc/installation.md)

## Using prepare-tutorial.sh to skip tutorials 

These tutorials are designed to be followed in sequence, but if you'd like to skip ahead, a script is supplied in the <code>granitic-examples</code> package which will set up a working project ready for you to follow the tutorial you're interested in.

Make sure you've downloaded the <code>granitic-examples</code> package:

<pre>
go get github.com/graniticio/granitic-examples
</pre>

Then (assuming you'd like to skip to tutorial 2) run:

<pre>
cd $GOPATH/src/github.com/graniticio/granitic-examples/tutorial
./prepare-tutorial.sh 2
</pre>

You'll now find a working Granitic project in <code>$GOHOME/src/granitic-tutorial/recordstore</code> in the correct state for starting tutorial 2

## Using an IDE with the tutorials

It is recommended you create your IDE project in <code>$GOHOME/src/granitic-tutorial</code>

## Notes for Windows users

The tutorials use UNIX conventions for file paths and environment variables. You will need to adapt the tutorials as you
go. Remember:

 * Replace / characters in paths with \
 * Replace $VARNAME with %VARNAME% when dealing with environment variables
 * Omit the leading <code>./</code> when running your compiled programmes (e.g. <code>service</code> rather than <code>./service</code>)
 * mkdir on Windows does not need a -p switch to create missing directories

## Tutorials

 1. [Fundamentals](001-fundamentals.md) - Learn more about Granitic and create a simple web service
 2. [Configuration](002-configuration.md) - How Granitic applications are configured