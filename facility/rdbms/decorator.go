// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/rdbms"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

type clientManagerDecorator struct {
	fieldNames []string
	manager    rdbms.RDBMSClientManager
}

func (cmd *clientManagerDecorator) OfInterest(component *ioc.Component) bool {

	result := false

	for _, f := range cmd.fieldNames {

		i := component.Instance

		if fieldPresent := reflecttools.HasFieldOfName(i, f); !fieldPresent {
			continue
		}

		targetFieldType := reflecttools.TypeOfField(i, f)
		managerType := reflect.TypeOf(cmd.manager)

		v := reflect.ValueOf(i).Elem().FieldByName(f)

		if managerType.AssignableTo(targetFieldType) && v.IsNil() {
			return true
		}
	}

	return result
}

func (cmd *clientManagerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	for _, f := range cmd.fieldNames {

		i := component.Instance

		if fieldPresent := reflecttools.HasFieldOfName(i, f); !fieldPresent {
			continue
		}

		targetFieldType := reflecttools.TypeOfField(i, f)
		managerType := reflect.TypeOf(cmd.manager)

		v := reflect.ValueOf(i).Elem().FieldByName(f)

		if managerType.AssignableTo(targetFieldType) && v.IsNil() {
			rc := reflect.ValueOf(i).Elem()
			rc.FieldByName(f).Set(reflect.ValueOf(cmd.manager))
		}
	}

}
