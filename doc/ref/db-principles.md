# Relational database principles

[Reference](README.md) | [Relational Databases](db-index.md)

---

## Insulation of application code from vendor specifics

Granitic has the concept of a [database provider](db-provider.md) to insulate your application code from vendor specific
details like connection management, driver code and auto-incremented ID recovery.

For each type of SQL database you connect to, you should only have to write a single, small component and
the rest of your application code will use interfaces and abstractions provided by Granitic for managing and executing
queries.

## Separation of Go code and queries

The [query manager](db-query.md) facility allows you to define your queries as templates in plain text files which
are populated at runtime. This makes your Go code cleaner, allows database specialists to more easily review and tune
your queries and allows queries to be modified without recompiling your application.

## Auto-injection of components

In the most common use case of an application accessing a single database, your application code just needs to [define
a variable with a standard type and name](db-provider.md) and Granitic will inject a [client manager](db-provider.md)
from which your code can request a client interface to start executing queries.

## Single interface

Your application will work with a [single interface](db-execution.md) combining templated query execution, raw query
query execution and transaction management.

## Minimised boilerplate

Methods on the [client interface](db-execution.md) that return data use a reflection-based technique to dynamically
map result rows to strongly typed structs that you define.

## More readable application code

The methods on the [client interface](db-execution.md) are named in such a way as the intent of your data access code is 
more clear.   

---
**Next**: [Database provider](db-provider.md)

**Prev**: [Relational databases](db-index.md)
