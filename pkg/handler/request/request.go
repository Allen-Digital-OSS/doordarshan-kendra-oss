package request

import (
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"error"`
}

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{validator: validator.New(validator.WithRequiredStructEnabled())}
}

func (rv *Validator) Validate(i interface{}) error {
	return rv.validator.Struct(i)
}
