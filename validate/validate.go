package validate

type ValidateContext struct {
	V           interface{}
	FieldName   string
	BoundFields []string
}

func (vc *ValidateContext) WasBound(code string) bool {

	if vc.BoundFields != nil {

		for _, f := range vc.BoundFields {
			if f == vc.FieldName {
				return true
			}
		}
	}

	return false
}

type Validator interface {
	Validate(vc *ValidateContext) (failcodes []string, err error)
}

func NewValidateContext(v interface{}) *ValidateContext {
	vc := new(ValidateContext)
	vc.V = v

	return vc
}
