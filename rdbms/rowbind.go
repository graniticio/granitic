// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"database/sql"
	"errors"
	"fmt"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
	"strconv"
)

// Used to extract the data from the results of a SQL query and inject the data into a target data structure.
type RowBinder struct {
}

/*
	BindRow takes results from a SQL query that has return zero rows or one row and maps the data into the
	target interface, which must be a pointer to a struct.

	If the query results contain zero rows, BindRow returns false, nil.

	If the query results contain one row, it is populated. See the GoDoc for BindRows for more detail.

*/
func (rb *RowBinder) BindRow(r *sql.Rows, t interface{}) (bool, error) {

	if r == nil{
		return false, errors.New("Nil sql.Rows supplied to BindRow")
	}

	if !rt.IsPointerToStruct(t) {

		if rt.IsPointer(t) {
			if r.Next() {
				return true, r.Scan(t)
			}

			return false, nil
		} else {
			m := fmt.Sprintf("Target must be a pointer to a struct or pointer. Is %T", t)
			return false, errors.New(m)
		}
	}

	if results, err := rb.BindRows(r, t); err != nil {
		return false, err
	} else {

		rs := len(results)

		if rs == 0 {
			return false, nil
		}

		if rs != 1 {
			m := fmt.Sprintf("BindRow: query returned %d rows, expected zero or one row.", rs)
			return false, errors.New(m)
		}

		tr := reflect.ValueOf(t).Elem()

		rr := reflect.ValueOf(results[0]).Elem()

		for i := 0; i < tr.NumField(); i++ {

			tr.Field(i).Set(rr.Field(i))

		}

	}

	return true, nil
}

/*
	BindRow takes results from a SQL query that has return zero rows or one row and maps the data into the
	instances of the target interface, which must be a pointer to a struct.

	If the query results contain zero rows, BindRow returns an empty slice of the target type

	If the query results contain one or more rows, an instance of the target type is created for each row. Each column
	in a row is mapped to a field in the target type by either:

	a) Finding a field whose name exactly matches the column name or alias.

	b) Finding a field with the 'column' struct tag with a value that exactly matches the column name or alias.

	A target field may be a bool, any native int/uint type, any native float type, a string or any of the
	Granitic nilable types.
*/
func (rb *RowBinder) BindRows(r *sql.Rows, t interface{}) ([]interface{}, error) {

	var err error
	var columnNames []string
	var targetScanners map[string]*scanner

	if r == nil {
		return nil, errors.New("Nil *sql.Rows supplied")
	}


	if !rt.IsPointerToStruct(t) {
		return nil, errors.New("Template must be a pointer to a struct.")
	}

	if columnNames, err = r.Columns(); err != nil {
		return nil, err
	}

	colCount := len(columnNames)

	targetScanners = rb.generateTargets(t)


	scanners := make([]interface{}, colCount)

	results := make([]interface{}, 0)

	matchedTargets := 0

	for i, cn := range columnNames {

		scanner := targetScanners[cn]

		if scanner == nil {

			m := fmt.Sprintf("No field available to receive column %s (no matching field name or 'column:' tag)", cn)
			return nil, errors.New(m)
		}

		scanners[i] = scanner
		matchedTargets++

	}

	if matchedTargets != colCount {
		m := fmt.Sprintf("Not all of the columns in the results could be matched to fields on the template.")
		return nil, errors.New(m)
	}

	for r.Next() {

		if err := r.Scan(scanners...); err != nil {
			return nil, err
		}

		if built, err := rb.buildAndPopulate(t, scanners);err != nil {
			return nil, err
		} else {
			results = append(results, built)
		}

	}

	return results, nil
}

func (rb *RowBinder) buildAndPopulate(t interface{}, scanners []interface{}) (r interface{}, err error) {

	r = reflect.New(reflect.TypeOf(t).Elem()).Interface()

	rv := reflect.ValueOf(r).Elem()

	for _, s := range scanners {

		v := s.(*scanner)

		f := rv.FieldByName(v.field)

		if v.val != nil {

			pv := reflect.ValueOf(v.val)

			defer func() {
				if r := recover(); r != nil {
					err = errors.New(fmt.Sprintf("Unable to set field %s with value of type %T\n", v.field, pv.Interface()))
				}
			}()


			f.Set(pv)
		}

	}

	return r, err

}

func (rb *RowBinder) generateTargets(t interface{}) (map[string]*scanner) {

	targets := make(map[string]*scanner)

	rv := reflect.ValueOf(t).Elem()
	rt := reflect.TypeOf(t).Elem()

	FieldLoop:
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		alias := f.Tag.Get("column")

		fv := rv.FieldByName(f.Name)

		if fv.CanSet() {

			s := new(scanner)
			s.field = f.Name
			s.kind = fv.Kind()

			if s.kind == reflect.Ptr {

				i := fv.Interface()

				switch i.(type) {

				case *types.NilableBool:
					s.nilable = NilBool
				case *types.NilableString:
					s.nilable = NilString
				case *types.NilableFloat64:
					s.nilable = NilFloat
				case *types.NilableInt64:
					s.nilable = NilInt
				default:
					//Ignore other fields
					continue FieldLoop
				}

			}

			if alias != "" {
				targets[alias] = s
			} else {
				targets[f.Name] = s

			}

		}
	}

	return targets
}

type nilableType int

const (
	Unset = iota
	NilBool
	NilString
	NilInt
	NilFloat
)

type scanner struct {
	kind    reflect.Kind
	field   string
	nilable nilableType
	val     interface{}
}

func (s *scanner) Scan(src interface{}) error {

	if b, found := src.([]byte); found {
		sv := string(b)

		if s.kind != reflect.String {
			return s.convert(sv)
		} else {
			s.val = sv
		}

	} else {
		s.val = src
	}

	return nil
}

func (s *scanner) convert(sv string) error {

	switch s.kind {
	case reflect.Ptr:
		return s.supportedStruct(sv)
	case reflect.Int:
		return s.toIntVal(sv, 0)
	case reflect.Int8:
		return s.toIntVal(sv, 8)
	case reflect.Int16:
		return s.toIntVal(sv, 16)
	case reflect.Int32:
		return s.toIntVal(sv, 32)
	case reflect.Int64:
		return s.toIntVal(sv, 64)
	case reflect.Uint:
		return s.toUintVal(sv, 0)
	case reflect.Uint8:
		return s.toUintVal(sv, 8)
	case reflect.Uint16:
		return s.toUintVal(sv, 16)
	case reflect.Uint32:
		return s.toUintVal(sv, 32)
	case reflect.Uint64:
		return s.toUintVal(sv, 64)
	case reflect.Float32:
		return s.toFloatVal(sv, 32)
	case reflect.Float64:
		return s.toFloatVal(sv, 64)
	case reflect.Bool:
		if b, err := strconv.ParseBool(sv); err != nil {
			return err
		} else {
			s.val = b
		}

	default:
		m := fmt.Sprintf("RowBinder: Unsupported type '%v' in target object. Value from database was: %s", s.kind, sv)
		return errors.New(m)

	}

	return nil
}

func (s *scanner) supportedStruct(sv string) error {

	switch s.nilable {

	case NilBool:

		if b, err := strconv.ParseBool(sv); err != nil {
			return err
		} else {
			s.val = types.NewNilableBool(b)
		}

	case NilInt:

		if err := s.toIntVal(sv, 64); err != nil {
			return err
		} else {
			s.val = types.NewNilableInt64(s.val.(int64))
		}

	case NilFloat:

		if err := s.toFloatVal(sv, 64); err != nil {
			return err
		} else {
			s.val = types.NewNilableFloat64(s.val.(float64))
		}

	case NilString:
		s.val = types.NewNilableString(sv)

	}

	return nil

}

func (s *scanner) toFloatVal(sv string, size int) error {
	if i, err := strconv.ParseFloat(sv, size); err != nil {
		return err
	} else {

		switch size {
		case 32:
			s.val = float32(i)
		case 64:
			s.val = float64(i)
		}

	}

	return nil
}

func (s *scanner) toIntVal(sv string, size int) error {

	if i, err := strconv.ParseInt(sv, 10, size); err != nil {
		return err
	} else {

		switch size {
		case 0:
			s.val = int(i)
		case 8:
			s.val = int8(i)
		case 16:
			s.val = int16(i)
		case 32:
			s.val = int32(i)
		case 64:
			s.val = int64(i)
		}

	}

	return nil
}

func (s *scanner) toUintVal(sv string, size int) error {

	if i, err := strconv.ParseUint(sv, 10, size); err != nil {
		return err
	} else {

		switch size {
		case 0:
			s.val = uint(i)
		case 8:
			s.val = uint8(i)
		case 16:
			s.val = uint16(i)
		case 32:
			s.val = uint32(i)
		case 64:
			s.val = uint16(i)
		}

	}

	return nil
}
