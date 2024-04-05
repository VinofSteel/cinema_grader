package validation

import (
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

var Validate *validator.Validate

func ValidateData(data interface{}) []ErrorResponse {
	var validationErrors []ErrorResponse

	errors := Validate.Struct(data)
	if errors != nil {
		for _, err := range errors.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field()
			elem.Tag = err.Tag()
			elem.Value = err.Value()
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}
