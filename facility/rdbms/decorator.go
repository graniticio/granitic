// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/rdbms"
	"github.com/graniticio/granitic/v2/reflecttools"
	"reflect"
)

type clientManagerDecorator struct {
	fieldNameManager map[string]rdbms.RdbmsClientManager
	log              logging.Logger
}

func (cmd *clientManagerDecorator) OfInterest(component *ioc.Component) bool {

	result := false

	for field, manager := range cmd.fieldNameManager {

		i := component.Instance

		if fieldPresent := reflecttools.HasFieldOfName(i, field); !fieldPresent {
			continue
		}

		targetFieldType := reflecttools.TypeOfField(i, field)
		managerType := reflect.TypeOf(manager)

		v := reflect.ValueOf(i).Elem().FieldByName(field)

		if managerType.AssignableTo(targetFieldType) && v.IsNil() {

			cmd.log.LogTracef("%s.%s needs a client manager", component.Name, field)

			return true
		}
	}

	return result
}

func (cmd *clientManagerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	for field, manager := range cmd.fieldNameManager {

		i := component.Instance

		if fieldPresent := reflecttools.HasFieldOfName(i, field); !fieldPresent {
			continue
		}

		targetFieldType := reflecttools.TypeOfField(i, field)
		managerType := reflect.TypeOf(manager)

		v := reflect.ValueOf(i).Elem().FieldByName(field)

		if managerType.AssignableTo(targetFieldType) && v.IsNil() {
			rc := reflect.ValueOf(i).Elem()
			rc.FieldByName(field).Set(reflect.ValueOf(manager))
		}
	}

}
