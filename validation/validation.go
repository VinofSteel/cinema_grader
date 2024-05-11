package validation

import (
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error        bool
	FailedField  string
	Tag          string
	ErrorMessage string
}

func firstAndLastToLower(s string) string {
	if len(s) == 0 {
		return s
	}

	firstRune, size := utf8.DecodeRuneInString(s)
	if firstRune == utf8.RuneError && size <= 1 {
		return s
	}

	lastRune, lastSize := utf8.DecodeLastRuneInString(s)
	if lastRune == utf8.RuneError && lastSize <= 1 {
		return s
	}

	firstLower := unicode.ToLower(firstRune)
	lastLower := unicode.ToLower(lastRune)

	// If the first and last characters are already lowercase, return the original string.
	if firstRune == firstLower && lastRune == lastLower {
		return s
	}

	return string(firstLower) + s[size:len(s)-lastSize] + string(lastLower)
}

func structValidation(validate *validator.Validate, data interface{}) []ErrorResponse {
	var validationErrors []ErrorResponse

	errors := validate.Struct(data)
	if errors != nil {
		for _, err := range errors.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = firstAndLastToLower(err.Field())
			elem.Tag = err.Tag()
			elem.Error = true

			switch err.Tag() {
			case "required":
				elem.ErrorMessage = fmt.Sprintf("The %s field is required.", firstAndLastToLower(err.Field()))
			case "password":
				elem.ErrorMessage = "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number."
			case "email":
				elem.ErrorMessage = "The email field needs to be a valid email."
			case "datetime":
				elem.ErrorMessage = fmt.Sprintf("The %s field needs to follow the YYYY-MM-DD format.", firstAndLastToLower(err.Field()))
			case "isadminuuid":
				elem.ErrorMessage = "The creatorId field needs to be a valid uuid that belongs to an admin user."
			case "validactorslice":
				elem.ErrorMessage = "The actors field needs to be a valid array that contains uuids of existing actors."
			case "isvaliduuid":
				elem.ErrorMessage = fmt.Sprintf("The %s field needs to be a valid uuid.", firstAndLastToLower(err.Field()))
			case "isvalidgrade":
				elem.ErrorMessage = fmt.Sprintf("The %s field needs to a float between 1.0 and 5.0, with only one decimal field.", firstAndLastToLower(err.Field()))

			}

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func ValidateData(c *fiber.Ctx, validate *validator.Validate, data interface{}) bool {
	if errors := structValidation(validate, data); len(errors) > 0 && errors[0].Error {
		log.Println("Errors while validating data in the ValidateData function...", errors)
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
