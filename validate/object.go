package validate

import "github.com/graniticio/granitic/types"

const (
	objOpRequiredCode = commonOpRequired
	objOpStopAllCode  = commonOpStopAll
)

type ObjectValidationOperation uint

const (
	ObjOpUnsupported = iota
	ObjOpRequired
)

type ObjectValidator struct {
	stopAll    bool
	codesInUse types.StringSet
	dependsOn  types.StringSet
}

func (ov *ObjectValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {
	return nil, nil
}

func (ov *ObjectValidator) StopAllOnFail() bool {
	return ov.stopAll
}

func (ov *ObjectValidator) CodesInUse() types.StringSet {
	return ov.codesInUse
}

func (ov *ObjectValidator) DependsOnFields() types.StringSet {
	return ov.DependsOnFields()
}
