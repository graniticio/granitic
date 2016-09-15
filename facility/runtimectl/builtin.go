package runtimectl

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"strconv"
	"strings"
)

const (
	fwArg = "fw"
	lcArg = "lc"
	rcArg = "rc"
)

func findLifecycleFilter(args map[string]string) (lifecycleFilter, error) {

	if args == nil || len(args) == 0 || args[lcArg] == "" {
		return all, nil
	}

	v := args[lcArg]

	return fromFilterArg(v)

}

func RuntimeCtlEnabled(ca *config.ConfigAccessor) bool {

	p := "Facilities.RuntimeCtl"

	if !ca.PathExists(p) {
		return false
	}

	b, _ := ca.BoolVal(p)

	return b

}

func includeRuntime(args map[string]string) (bool, error) {
	return boolArg(args, rcArg)
}

func showBuiltin(args map[string]string) (bool, error) {
	return boolArg(args, fwArg)
}

func boolArg(args map[string]string, n string) (bool, error) {

	if args == nil || len(args) == 0 || args[n] == "" {
		return false, nil
	}

	v := args[n]

	if choice, err := strconv.ParseBool(v); err == nil {
		return choice, nil
	} else {

		m := fmt.Sprintf("Value of %s argument cannot be interpreted as a bool", n)

		return false, errors.New(m)
	}

	return false, nil

}

func matchesFilter(f lifecycleFilter, i interface{}) bool {

	switch f {
	case all:
		return true

	case start:
		_, found := i.(ioc.Startable)
		return found

	case stop:
		_, found := i.(ioc.Stoppable)
		return found

	case suspend:
		_, found := i.(ioc.Suspendable)
		return found

	}

	return true
}

type lifecycleFilter int

const (
	all = iota
	stop
	start
	suspend
)

func fromFilterArg(arg string) (lifecycleFilter, error) {

	s := strings.ToLower(arg)

	switch s {
	case "", "all":
		return all, nil
	case "stop":
		return stop, nil
	case "start":
		return start, nil
	case "suspend":
		return suspend, nil
	}

	m := fmt.Sprintf("%s is not a recognised lifecycle filter (all, stop, start, suspend)", arg)

	return all, errors.New(m)

}

func isFramework(c *ioc.Component) bool {
	return strings.HasPrefix(c.Name, instance.FrameworkPrefix)
}
