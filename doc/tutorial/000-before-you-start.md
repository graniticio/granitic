# Tutorial - Before you start

Make sure you have followed the Granitic [installation instructions](../installation.md)

## Starting points for tutorials 

These tutorials are designed to be followed in sequence, but if you'd like to skip ahead, 
a GitHub repository is available containing working start points for each
of the tutorials. Different versions of the tutorial source code are provided for those
who prefer to work with JSON configuration files and those who prefer YAML.

The step-by-step tutorials that follow all use JSON as the example format.

### Checking out the tutorials repository

You can clone the tutorials repo to any location on your local machine. In the example
below we have cloned the repo to `~/grnc-tutorial`

<pre>
cd ~
git clone https://github.com/graniticio/tutorial.git grnc-tutorial
</pre>


### Using an IDE with the tutorials

It is recommended you use your IDE to open either the `json` or `yaml` folder in the tutorials
repo you checked out above. This means you will have all of the tutorials visible
in a single project.

## Notes for Windows users

The tutorials use UNIX conventions for file paths and environment variables. You will need to adapt the tutorials as you
go. Remember:

 * Replace / characters in paths with \
 * Replace $VARNAME with %VARNAME% when dealing with environment variables
 * Omit the leading <code>./</code> when running your compiled programmes (e.g. <code>service</code> rather than <code>./service</code>)
 * mkdir on Windows does not need a -p switch to create missing directories

## Tutorials

The [first tutorial](001-fundamentals.md) will show you how to build a simple web-service using Granitic and Go