package initializers

import (
	"log"
	"regexp"
	"sync"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var UserModel models.UserModel
var ActorModel models.ActorModel

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

func uuidValidation(fl validator.FieldLevel) bool {
	idField := fl.Field().String()

	_, err := uuid.Parse(idField)
	if err != nil {
		log.Println("Error parsing uuid:", err)
		return false
	}

	return true
}

func actorsUuidSliceValidation(fl validator.FieldLevel) bool {
	db := NewDatabaseConn()
	defer db.Close()

	field := fl.Field()
	actorsField := field.Interface().([]string)
	if len(actorsField) == 0 {
		log.Println("Actors field cannot be empty when creating a movie")
		return false
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(actorsField))
	for _, actorID := range actorsField {
		wg.Add(1)
		go func(actorID string) {
			defer wg.Done()
			uuid, err := uuid.Parse(actorID)
			if err != nil {
				log.Println("Error parsing actor uuid:", err)
				errCh <- err
				return
			}

			_, err = ActorModel.GetActorById(db, uuid)
			if err != nil {
				log.Println("Error getting actor by id when validating actor uuids:", err)
				errCh <- err
				return
			}
		}(actorID)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error in actor verification goroutine: %v", err)
			return false
		}
	}

	return true
}

func gradeValidation(fl validator.FieldLevel) bool {
	grade, ok := fl.Field().Interface().(float64)
	if !ok {
		return false
	}

	const epsilon = 0.000001 // Tolerance for floating-point comparison

	if grade < 1.0-epsilon || grade > 5.0+epsilon {
		return false
	}

	return true
}

func NewValidator() *validator.Validate {
	// Initializing a single instance of the validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Validator custom functions
	validate.RegisterValidation("password", passwordValidation)
	validate.RegisterValidation("isadminuuid", adminUuidValidation)
	validate.RegisterValidation("validactorslice", actorsUuidSliceValidation)
	validate.RegisterValidation("isvaliduuid", uuidValidation)
	validate.RegisterValidation("isvalidgrade", gradeValidation)

	return validate
}
