package validation

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error        bool
	FailedField  string
	Tag          string
	ErrorMessage string
}

var Validate *validator.Validate

func StructValidation(data interface{}) []ErrorResponse {
	var validationErrors []ErrorResponse

	errors := Validate.Struct(data)
	if errors != nil {
		for _, err := range errors.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = strings.ToLower(strings.ToLower(err.Field()))
			elem.Tag = err.Tag()
			elem.Error = true

			switch err.Tag() {
			case "required":
				elem.ErrorMessage = fmt.Sprintf("The %s field is required.", strings.ToLower(err.Field()))
			case "password":
				elem.ErrorMessage = "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number."
			case "email":
				elem.ErrorMessage = "The email field needs to be a valid email."
			case "datetime":
				elem.ErrorMessage = fmt.Sprintf("The %s field needs to follow the YYYY-MM-DD format.", strings.ToLower(err.Field()))
			}

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func ValidateData(c *fiber.Ctx, data interface{}) bool {
	log.Println("Executing ValidateData function...")
	if errors := StructValidation(data); len(errors) > 0 && errors[0].Error {
		errMap := make(map[string]string)

		for _, err := range errors {
			errMap[err.FailedField] = err.ErrorMessage
		}

		c.Status(fiber.ErrBadRequest.Code).JSON(struct {
			Message string            `json:"message"`
			Errors  map[string]string `json:"errors"`
		}{
			Message: "Validation failed",
			Errors:  errMap,
		},
		)

		return false
	}

	return true
}
