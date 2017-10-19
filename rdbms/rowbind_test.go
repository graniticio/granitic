package rdbms

import (
	"testing"
	"reflect"
	"github.com/graniticio/granitic/test"
	"fmt"
	"github.com/graniticio/granitic/types"
)

func TestScan(t *testing.T) {

	s := new(scanner)

	err := s.Scan("string")
	test.ExpectNil(t, err)

	s.kind = reflect.String
	err = s.Scan("string")
	test.ExpectNil(t, err)

	s.kind = reflect.Bool
	err = s.Scan([]byte("true"))

	test.ExpectNil(t, err)
	b, f := s.val.(bool)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, b, true)

	s.kind = reflect.Bool
	err = s.Scan([]byte("false"))

	test.ExpectNil(t, err)
	b, f = s.val.(bool)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, b, false)

	s.kind = reflect.Bool
	err = s.Scan([]byte("xxx"))

	test.ExpectNotNil(t, err)

	for _, k := range []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64} {

		s.kind = k
		err = s.Scan([]byte("123"))

		test.ExpectNil(t, err)

		err = s.Scan([]byte("xx"))

		test.ExpectNotNil(t, err)

	}

	for _, k := range []reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64} {

		s.kind = k
		err = s.Scan([]byte("123"))

		sv := fmt.Sprintf("%d", s.val)

		test.ExpectNil(t, err)


		test.ExpectString(t, sv, "123")

		err = s.Scan([]byte("-1"))

		test.ExpectNotNil(t, err)

	}

	for _, k := range []reflect.Kind{reflect.Float32, reflect.Float64} {

		s.kind = k
		err = s.Scan([]byte("123.1"))

		sv := fmt.Sprintf("%v", s.val)

		test.ExpectNil(t, err)



		test.ExpectString(t, sv, "123.1")

		err = s.Scan([]byte("xx"))

		test.ExpectNotNil(t, err)

	}


	s.kind = reflect.Struct
	err = s.Scan([]byte("xx"))
	test.ExpectNotNil(t, err)
}

func TestNilable(t *testing.T) {

	s := new(scanner)
	s.kind = reflect.Ptr
	s.nilable = NilBool


	err := s.Scan([]byte("true"))
	test.ExpectNil(t, err)


	nb, f := s.val.(*types.NilableBool)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, f, nb.IsSet())
	test.ExpectBool(t, f, nb.Bool())

	err = s.Scan([]byte("false"))
	test.ExpectNil(t, err)

	err = s.Scan([]byte("xx"))
	test.ExpectNotNil(t, err)


	s.nilable = NilFloat

	err = s.Scan([]byte("32.2"))
	test.ExpectNil(t, err)


	nf, f := s.val.(*types.NilableFloat64)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, f, nf.IsSet())
	test.ExpectFloat(t, nf.Float64(), float64(32.2))


	s.nilable = NilInt

	err = s.Scan([]byte("32"))
	test.ExpectNil(t, err)


	ni, f := s.val.(*types.NilableInt64)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, f, ni.IsSet())
	test.ExpectInt(t, int(ni.Int64()), int(32))


	s.nilable = NilString

	err = s.Scan([]byte("XX"))
	test.ExpectNil(t, err)


	ns, f := s.val.(*types.NilableString)

	test.ExpectBool(t, f, true)
	test.ExpectBool(t, f, nb.IsSet())
	test.ExpectString(t, ns.String(), "XX")

}