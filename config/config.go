// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package config provides functionality for working with configuration files and command line arguments to a Granitic application.

Grantic uses JSON files to store component definitions (declarations of, and relationships between, components to
run in the IoC container) and configuration (variables used by IoC components that may vary between environments and settings
for Grantic's built-in facilities). A definition of the use and syntax of these files are outside of the scope of a GoDoc page,
but are described in detail at https://granitic.io/ref/component-definition-files and https://granitic.io/ref/configuration-files

This package defines functionality for loading a JSON file (from a filesystem or via HTTP) and merging multiple files into
a single view. This is a key concept in Granitic.

Given a folder of configuration files called conf:

	conf/x.json
	conf/sub/a.json
	conf/sub/b.json

starting a Grantic application with:

	-c http://example.com/base.json,conf,http://example.com/myinstance.json

The following will take place. Firstly the files would be expanded into a flat list of paths/URIs

	http://example.com/base.json
	conf/sub/a.json
	conf/sub/b.json
	conf/x.json
	http://example.com/myinstance.json

The the files will be merged together from left, using the the first file as a base. In this example,  http://example.com/base.json
and conf/sub/a.json will be merged together, then result of that merge will be merged with conf/sub/b.json and so on.

For named fields (in a JSON object/map), the process of merging is fairly obvious. When merging files A and B, a field that
is defined in both files will have the value of the field used in file B in the merged output. For example,

	a.json

	{
		"database": {
			"host": "localhost",
			"port": 3306,
			"flags": ["a", "b", "c"]
		}
	}

and

	b.json

	{
		"database": {
			"host": "remotehost",
			"flags": ["d"]
		}
	}

woud merge to:

	{
		"database": {
			"host": "remotehost",
			"port": 3306,
			"flags": ["d"]
		}
	}

The merging of configuration files occurs exactly above, but when component definition files are merged, arrays are joined, not overwritten.
For example:

	{ "methods": ["GET"] }

merged with;

	{ "methods": ["POST"] }

would result in:

	{ "methods": ["GET", "POST"] }

Another core concept used by the types in this package is a config path. This is the absolute path to field in the
eventual merged configuration file with a dot-delimited notation. E.g "database.host".
*/
package config

// JSONPathSeparator is the character used to delimit paths to config values.
const JSONPathSeparator string = "."

// Used by code that needs to know what type of JSON data structure resides at a particular path
// before operating on it.
/*const (
	Unset       = -2
	JSONUnknown = -1
	JSONString  = 1
	JSONArray   = 2
	JSONMap     = 3
	JSONBool    = 4
)

// JSONType determines the apparent JSONType of the supplied Go interface.
func JSONType(value interface{}) int {

	switch value.(type) {
	case string:
		return JSONString
	case map[string]interface{}:
		return JSONMap
	case bool:
		return JSONBool
	case []interface{}:
		return JSONArray
	default:
		return JSONUnknown
	}
}*/

// MissingPathError indicates that the a problem was caused by there being no value at the supplied
// config path
type MissingPathError struct {
	message string
}

func (mp MissingPathError) Error() string {
	return mp.message
}
