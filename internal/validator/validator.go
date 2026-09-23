package validator

import (
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func FormatValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errs[e.Field()] = e.Tag()
		}
	}
	return errs
}