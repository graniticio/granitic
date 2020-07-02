# Query Manager
[Reference](README.md) | [Facilities](fac-index.md)

---

Enabling the QueryManager facility allows you to store query templates in plain text files instead of
your application's Go code.

Any type of query can be stored in the templates, but the facility optionally provides additional support for SQL queries
and integrates closely with [RDBMS Access facility](fac-rdbms.md).

Usage of this facility is covered in detail in the [query management section](db-query.md) of the reference manual. The
remainder of this section covers the configuration and customisation of the facility.

## Enabling

The QueryManager facility is _disabled_ by default. To enable it, you must set the following in your configuration

```json
{
  "Facilities": {
    "QueryManager": true
  }
}
``` 

## Configuration

The default configuration for this facility can be found in the Granitic source under `facility/config/querymanager.json`
and is:

```json
{
  "QueryManager":{
    "TemplateLocation": "resource/queries",
    "QueryIDPrefix": "ID:",
    "TrimIDWhiteSpace": true,
    "VarMatchRegEx": "\\$\\{([^\\}]*)\\}",
    "NewLine": "\n",
    "CreateDefaultValueProcessor": true,
    "ProcessorName": "CONFIGURABLE",
    "ElementSeparator": ", ",
    "ValueProcessors": {
      "Configurable": {
        "WrapStrings": true,
        "StringWrapWith": "'"
      },
      "SQL": {
        "BoolFalse": 0,
        "BoolTrue": 1
      }
    }
  }
}
```

## Template location

Alter the value of `QueryManager.TemplateLocation` to change where Granitic looks for query template files. Note
that relative directories will be relative to the working directory in which your application is running.

Granitic will recursively search all child directories and any file found will be parsed for queries.

## Template parsing

### Query IDs

`QueryManager.QueryIDPrefix` defines prefix the lines in a query template file that signify the start of a new query
and define the unique ID for that query. For example, a line starting `ID:ALL_ARTISTS` would indicate that the 
following text represents a query that has the ID `ALL_ARTISTS`.

`QueryManager.TrimIDWhiteSpace` defines whether query IDs should be trimmed for whitespace.

## Line endings

`QueryManager.NewLine` is unused and will be removed in a future version of Granitic

## Variables

`QueryManager.VarMatchRegEx` defines the format of variables in template files (by default they are `${VarName}`). If
you change this, your regex must include a single capture group which isolates the name of the variable.

`QueryManager.VarMatchRegEx`



---
**Next**: [RDBMS integration](fac-rdbms.md)

**Prev**: [XML Web Services](fac-xml-ws.md)