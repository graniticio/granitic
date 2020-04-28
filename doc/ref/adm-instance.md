# Instance identification
[Reference](README.md) | [Administration](adm-index.md)

---

If you are running multiple instances of the same application, either on a single VM/server/container, or across multiple
hosts, it is often useful to assign each instance a unique ID.

You can provide an _instance ID_, to each instance of your application via a command line argument which is then made 
available to your application components if they need it.

## Providing an ID via the command line

Start your application with the `-i` argument. E.g.:

```shell script
myapp -i myid
```

You will see a message similar to:

```shell script
29/Oct/2019:11:01:08 Z INFO  [grncInit] Instance ID: myid
```

## Accessing the ID from your components

If your component implements [instance.Receiver](https://godoc.org/github.com/graniticio/granitic/instance#Receiver)
(e.g. implements the method `RegisterInstanceID(*instance.Identifier)`), the instance ID will be passed in to that
method before the [container starts your components](ioc-lifecycle.md).

The instance ID will also be injected into your application's [merged configuration](cfg-merging.md) and can be found at 
the config path ```System.InstanceID```

---
**Next**: [System configuration](adm-system.md)

**Prev**: [Remote configuration](adm-remote.md)