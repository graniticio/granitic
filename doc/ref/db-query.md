# Query Management

[Reference](README.md) | [Relational Databases](db-index.md)

---

Granitic includes the [query manager facility](fac-query.md) which allows your application to store database queries as templates in plain text files,
rather than embedding them in your source code.

The facility is highly customisable and can be used to manage queries for any queryable data source (search engine,
document database, etc), but this section of the documentation covers the most common use case of managing queries to be 
executed against a relational database.

Refer to the [facility documentation](fac-query.md) for more information on how to configure and customise.

## Query files

Once you have enabled the facility, Granitic will expect to find template files in the folder `resource/queries` - this
path is relative to wherever you start your application from. You can change this path by setting 
`QueryManager.TemplateLocation`.

Any files in that folder, or sub-folders, will be parsed as potential query template files. Each file can contain
zero to n queries and look like:

```sql
ID:CREATE_ARTIST

INSERT INTO artist(
  name,
  first_active
) VALUES (
  ${!Name},
  ${FirstYearActive}
)


ID:ARTIST_BY_ID

SELECT
  name AS Name
FROM
  artist
WHERE
  id = ${ID}
``` 

### Query IDs

A line that starts `ID:` indicates the start of a new query with the string after the colon becoming the query ID (or QID)
for that query. The example above defines two queries: `CREATE_ARTIST` and `ARTIST_BY_ID`

### Query template

All text after a query ID line until the next query ID line (or the end of the file) is consider to be the query template.
Lines that only contain whitespace will be discarded.

### Fragments

If your query template does not contain any variables, it is considered a fragment. Application
code using the [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#Client)
can recover these text fragments directly.


## Variables

Strings of the form `${Name}` or `${!Name}` are variables that will be substituted with values passed in during 
[query execution](db-execution.md). The `!` marks a variable as required - executing the query without providing a value
for a required variable will result in an error.

Values are provided either as a map or as struct. If the value source is a map, Granitic looks for a key with the exact
name of the variable (e.g `FirstYearActive` not `firstYearActive`). If the source is a struct, Granitic expects a field to be present that
shares the exact name of the variable (e.g `FirstYearActive` not `firstYearActive`) or for a field to have the
`dbparam` [struct tag](https://www.digitalocean.com/community/tutorials/how-to-use-struct-tags-in-go) set the 
exact name of the variable:

```go
type ArtistQuery struct {
  StartYear `dbparam:"FirstYearActive"`  
}
```

### Supported types for variable substitution

The types that you may use as values for variables depends on the value processor (see below) that the query manager is
configured to use. For more information on value processors, refer to the [facility documentation](fac-query.md)

Arrays of supported types will be rendered as comma de-limited lists.

## SQL mode

To enable SQL mode for the query manager, set `QueryManager.ProcessorName` to `"SQL"` in your configuration. This means
that when a query is generated from a template:

  * Unset (and unrequired) variables will be replaced with `null`
  * String variables will be wrapped in single quotes
  * bool variables will be replaced by 0 or 1 (configurable)

## Limitations

The current implementation of the query manager facility is only intended to support static queries. Dynamic
query construction (with the use of conditional statements) is not supported.

It is possible to avoid embedding queries in your code by:

  * Creating templates for fragments of queries and combining them at runtime
  * Making use of the `FindFragment` and `RegisterTempQuery` methods on the [rdbms.Client](https://godoc.org/github.com/graniticio/granitic/v2/rdbms#Client)
  interface made available to your code.


---
**Next**: [Executing queries](db-query.md)

**Prev**: [Database provider](db-provider.md)
