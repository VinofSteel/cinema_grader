package initializers

import (
	"log"
	"regexp"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var UserModel models.UserModel

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

func adminUuidValidation(fl validator.FieldLevel) bool {
	db := NewDatabaseConn()
	defer db.Close()

	idField := fl.Field().String()

	uuid, err := uuid.Parse(idField)
	if err != nil {
		log.Println("Error parsing admin uuid:", err)
		return false
	}

	userResponse, err := UserModel.GetUserById(db, uuid)
	if err != nil {
		log.Println("Error getting user by id when validating admin uuid:", err)
		return false
	}

	if !userResponse.IsAdm {
		log.Println("Valid and existing user uuid was passed in validation, but user isn't admin", err)
		return false
	}

	return true
}

func NewValidator() *validator.Validate {
	// Initializing a single instance of the validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Validator custom functions
	validate.RegisterValidation("password", passwordValidation)
	validate.RegisterValidation("isAdminUuid", adminUuidValidation)

	return validate
}
