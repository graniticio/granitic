package validate

import (
	"errors"
	"fmt"
)

const NoLimit = -1

const (
	stringOpTrimCode     = "TRIM"
	stringOpHardTrimCode = "HARDTRIM"
	stringOpLenCode      = "LEN"
	stringOpInCode       = "IN"
	stringOpExtCode      = "EXT"
	stringOpOptionalCode = "OPT"
	stringOpBreakCode    = "BREAK"
	stringOpRegCode      = "REG"
)

type StringValidationOperation uint

const (
	StringOpUnsupported = iota
	StringOpTrim
	StringOpHardTrim
	StringOpLen
	StringOpIn
	StringOpExt
	StringOpOptional
	StringOpBreak
	StringOpReg
)

type StringValidator struct {
	DefaultErrorcode string
	operations       []*stringOperation
	minLen           int
	maxLen           int
}

func (sv *StringValidator) Break() *StringValidator {

	o := new(stringOperation)
	o.OpType = StringOpBreak

	return sv

}

func (sv *StringValidator) Length(min, max int, code ...string) *StringValidator {

	sv.minLen = min
	sv.maxLen = max

	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.ErrCode = ec

	fmt.Printf("Adding %d %d, %s\n", min, max, ec)

	sv.addOperation(o)

	return sv

}

func (sv *StringValidator) addOperation(o *stringOperation) {
	if sv.operations == nil {
		sv.operations = make([]*stringOperation, 0)
	}

	sv.operations = append(sv.operations, o)
}

func (sv *StringValidator) Operation(c string) (StringValidationOperation, error) {
	switch c {
	case stringOpTrimCode:
		return StringOpTrim, nil
	case stringOpHardTrimCode:
		return StringOpHardTrim, nil
	case stringOpLenCode:
		return StringOpLen, nil
	case stringOpInCode:
		return StringOpIn, nil
	case stringOpExtCode:
		return StringOpExt, nil
	case stringOpOptionalCode:
		return StringOpOptional, nil
	case stringOpBreakCode:
		return StringOpBreak, nil
	case stringOpRegCode:
		return StringOpReg, nil
	}

	m := fmt.Sprintf("Unsupported string validation operation %s", c)
	return StringOpUnsupported, errors.New(m)

}

func (sv *StringValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		return v[0]
	} else {
		return sv.DefaultErrorcode
	}

}

type stringOperation struct {
	OpType  StringValidationOperation
	ErrCode string
}
