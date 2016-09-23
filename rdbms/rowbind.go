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

type RowBinder struct {
}

func (rb *RowBinder) BindRow(r *sql.Rows, t interface{}) error {
	return nil
}

func (rb *RowBinder) BindRows(r *sql.Rows, t interface{}) ([]interface{}, error) {

	var err error
	var columnNames []string
	var targetScanners map[string]*scanner

	if !rt.IsPointerToStruct(t) {
		return nil, errors.New("Template must be a pointer to a struct.")
	}

	if columnNames, err = r.Columns(); err != nil {
		return nil, err
	}

	colCount := len(columnNames)

	if targetScanners, err = rb.generateTargets(t); err != nil {
		return nil, err
	}

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

		results = append(results, rb.buildAndPopulate(t, scanners))

	}

	return results, nil
}

func (rb *RowBinder) buildAndPopulate(t interface{}, scanners []interface{}) interface{} {

	r := reflect.New(reflect.TypeOf(t).Elem()).Interface()

	rv := reflect.ValueOf(r).Elem()

	for _, s := range scanners {

		v := s.(*scanner)

		f := rv.FieldByName(v.field)

		if v.val != nil {
			f.Set(reflect.ValueOf(v.val))
		}

	}

	return r

}

func (rb *RowBinder) generateTargets(t interface{}) (map[string]*scanner, error) {

	targets := make(map[string]*scanner)

	rv := reflect.ValueOf(t).Elem()
	rt := reflect.TypeOf(t).Elem()

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

				switch t := i.(type) {

				case *types.NilableBool:
					s.nilable = NilBool
				case *types.NilableString:
					s.nilable = NilString
				case *types.NilableFloat64:
					s.nilable = NilFloat
				case *types.NilableInt64:
					s.nilable = NilInt
				default:
					m := fmt.Sprintf("Unsupported type %T on target objects.", t)
					return nil, errors.New(m)

				}

			}

			if alias != "" {
				targets[alias] = s
			} else {
				targets[f.Name] = s

			}

		}
	}

	return targets, nil
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
