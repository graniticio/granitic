package validate

import (
	"errors"
	"fmt"
)

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
}

func (sv *StringValidator) Break() *StringValidator {

	o := new(stringOperation)
	o.OpType = StringOpBreak

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

type stringOperation struct {
	OpType  StringValidationOperation
	ErrCode string
}
