package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error        bool
	FailedField  string
	Tag          string
	ErrorMessage string
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
			elem.Error = true

			switch err.Tag() {
			case "required":
				elem.ErrorMessage = fmt.Sprintf("The %s field is required.", err.Field())
			case "password":
				elem.ErrorMessage = "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number."
			case "email":
				elem.ErrorMessage = "The email field needs to be a valid email"
			case "datetime":
				elem.ErrorMessage = fmt.Sprintf("The %s field needs to follow the YYYY-MM-DD format.", err.Field())
			}

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}
