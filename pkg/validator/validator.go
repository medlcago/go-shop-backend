package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	validate *validator.Validate
}

func New(v *validator.Validate) *Validator {
	return &Validator{
		validate: v,
	}
}

func (v *Validator) Validate(out any) error {
	return v.validate.Struct(out)
}
