package initializers

import (
	"regexp"

	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/go-playground/validator/v10"
)

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
	validation.Validate = validator.New(validator.WithRequiredStructEnabled())

	// Validator custom functions
	validation.Validate.RegisterValidation("password", passwordValidation)
}