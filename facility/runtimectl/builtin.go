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

func findLifecycleFilter(args map[string]string) (ioc.LifecycleSupport, error) {

	if args == nil || len(args) == 0 || args[lcArg] == "" {
		return ioc.None, nil
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

func operateOnFramework(args map[string]string) (bool, error) {
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

func fromFilterArg(arg string) (ioc.LifecycleSupport, error) {

	s := strings.ToLower(arg)

	switch s {
	case "", "all":
		return ioc.None, nil
	case "stop":
		return ioc.CanStop, nil
	case "start":
		return ioc.CanStart, nil
	case "suspend":
		return ioc.CanSuspend, nil
	}

	m := fmt.Sprintf("%s is not a recognised lifecycle filter (all, stop, start, suspend)", arg)

	return ioc.None, errors.New(m)

}

func matchesFilter(f ioc.LifecycleSupport, i interface{}) bool {

	switch f {
	case ioc.None:
		return true

	case ioc.CanStart:
		_, found := i.(ioc.Startable)
		return found

	case ioc.CanStop:
		_, found := i.(ioc.Stoppable)
		return found

	case ioc.CanSuspend:
		_, found := i.(ioc.Suspendable)
		return found

	}

	return true
}

func isFramework(c *ioc.Component) bool {
	return strings.HasPrefix(c.Name, instance.FrameworkPrefix)
}
