// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package runtimectl

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/types"
	"sort"
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

// Enabled checks to see if the RuntimeCtl facility is enabled in configuration.
func Enabled(ca *config.Accessor) bool {

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

// OperateOnFramework returns true if framework components should be included in whatever command is being invoked.
func OperateOnFramework(args map[string]string) (bool, error) {
	return boolArg(args, fwArg)
}

func boolArg(args map[string]string, n string) (bool, error) {

	if args == nil || len(args) == 0 || args[n] == "" {
		return false, nil
	}

	v := args[n]

	if choice, err := strconv.ParseBool(v); err == nil {
		return choice, nil
	}

	return false, fmt.Errorf("value of %s argument cannot be interpreted as a bool", n)
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

func filteredComponents(cc *ioc.ComponentContainer, ls ioc.LifecycleSupport, of ownershipFilter, nameSort bool, ex ...string) []*ioc.Component {

	var base []*ioc.Component

	var exclude types.StringSet

	if ex == nil {
		exclude = types.NewEmptyOrderedStringSet()
	} else {
		exclude = types.NewOrderedStringSet(ex)
	}

	switch ls {
	case ioc.None:
		base = cc.AllComponents()
	default:
		base = cc.ByLifecycleSupport(ls)
	}

	filtered := make([]*ioc.Component, 0)

	for _, bc := range base {
		if exclude.Contains(bc.Name) {
			continue
		}

		switch of {
		case FrameworkOwned:
			if !isFramework(bc) {
				continue
			}
		case ApplicationOwned:
			if isFramework(bc) {
				continue
			}
		}

		filtered = append(filtered, bc)

	}

	if nameSort {
		sort.Sort(ioc.ByName{Components: filtered})
	}

	return filtered
}

func isFramework(c *ioc.Component) bool {
	return strings.HasPrefix(c.Name, instance.FrameworkPrefix)
}
