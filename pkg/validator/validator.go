package validator

import "github.com/go-playground/validator/v10"

type Validator interface {
	Validate(out any) error
}

type structValidator struct {
	validate *validator.Validate
}

func New(v *validator.Validate) *structValidator {
	return &structValidator{
		validate: v,
	}
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func (v *structValidator) Instance() *validator.Validate {
	return v.validate
}
