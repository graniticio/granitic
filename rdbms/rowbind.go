package rdbms

import (
	"database/sql"
	"errors"
	"fmt"
	rt "github.com/graniticio/granitic/reflecttools"
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

	if !rt.IsPointerToStruct(t) {
		return nil, errors.New("Template must be a pointer to a struct.")
	}

	if columnNames, err = r.Columns(); err != nil {
		return nil, err
	}

	colCount := len(columnNames)

	targetScanners := rb.generateTargets(t)
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
		f.Set(reflect.ValueOf(v.val))

	}

	return r

}

func (rb *RowBinder) generateTargets(t interface{}) map[string]*scanner {

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

			if alias != "" {
				targets[alias] = s
			} else {
				targets[f.Name] = s

			}

		}
	}

	return targets
}

type scanner struct {
	kind  reflect.Kind
	field string

	val interface{}
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

	default:
		m := fmt.Sprintf("RowBinder: Unsupported type '%v' in target object.", s.kind)
		return errors.New(m)

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
			s.val = int16(i)
		}

	}

	return nil
}
