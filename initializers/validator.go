package initializers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func passwordValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasSymbolRegex    = regexp.MustCompile(`[^a-zA-Z0-9]`)
		hasUppercaseRegex = regexp.MustCompile(`[A-Z]`)
		hasNumberRegex    = regexp.MustCompile(`[0-9]`)
	)

	return hasSymbolRegex.MatchString(password) && hasUppercaseRegex.MatchString(password) && hasNumberRegex.MatchString(password)
}

func InitializeValidator() {
	// Initializing a single instance of the validator
	Validate = validator.New(validator.WithRequiredStructEnabled())

	// Validator custom functions
	Validate.RegisterValidation("password", passwordValidation)
}
