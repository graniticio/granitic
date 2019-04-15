# Creating a Granitic project
[Reference](README.md) | [Granitic Projects](gpr-index.md)

---

## The grnc-project tool

Granitic includes a command line utility called `grnc-project` that automates the creation of new Granitic projects 
that can be immediately built and started. You can build this tool by running:

```
  go install github.com/graniticio/granitic/v2/cmd/grnc-project
```

As long as your `$GOPATH/bin` folder is in your `$PATH`, you will be able to run this tool from any folder

## Creating a project

Granitic fully supports Go modules. You can create your application project in any folder by running:

```
  grnc-project your-project-name
```

This will create a folder called `your-project-name` with a skeleton entry point file (`service.go`), `go.mod` file,
component definition file and configuration file.

### Module name

The Go module name assigned in the generated `go.mod` file is, by default the same as your project name. You can 
override this by running:

```
  grnc-project your-project-name module-name
```
---
**Next**: [Building and running your application](gpr-build.md) 

**Prev**: [Anatomy of a Granitic project](grp-anatomy.md)