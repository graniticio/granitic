package rdbms

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

type ClientManagerDecorator struct {
	fieldNames []string
	manager    RdbmsClientManager
}

func (cmd *ClientManagerDecorator) OfInterest(component *ioc.Component) bool {

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

func (cmd *ClientManagerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

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
