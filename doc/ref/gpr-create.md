# Creating a Granitic project

## The grnc-project tool

Granitic includes a command line utility called `grnc-project` that automates the creation of new Granitic projects that can be immediately built and started. You can build this tool by running:

```
  go install github.com/graniticio/granitic/cmd/grnc-project
```

As long as your `$GOPATH/bin` folder is in your `$PATH`, you will be able to run this tool from any folder

## Creating a project

Assuming that you are following the traditional Go approach of creating your projects within your GitHub account, you can create a new project by running:

```
  cd $GOPATH/github.com/yourAccount
  grnc-project your-project-name
```

This will create a folder called `your-project-name` with a skeleton entry point file (`service.go`), component definition file and configuration file.

### Editing the entry point file

If you open your newly created `service.go` file, you will see a line like:

```go
  import "./bindings"  //Change to a non-relative path if you want to use 'go install'
```

Which is a workaround to allow projects that are created outside of the normal `$GOHOME/src/github.com` location to compile with `go build`. This workaround is not compatible with `go install`, however. In the example above you should change this line to:

```go
  import "github.com/yourAccount/your-project-name/bindings"
```

You can avoid this by providing your full package name as a second argument to `grnc-project`:

```
  grnc-project your-project-name github.com/yourAccount/your-project-name
```

## Next

[Building and running your application](gpr-build.md)