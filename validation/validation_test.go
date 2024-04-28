package validation

import (
	"log"
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/tests"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var validate *validator.Validate
var userModel models.UserModel
var adminId string

func TestMain(m *testing.M) {
	// Setup
    var err error
    testDb, err := tests.Setup()
    if err != nil {
        log.Fatalf("Error setting up tests: %v", err)
    }

    os.Setenv("PGDATABASE", testDb)
	
    // Validator setup
	validate = initializers.NewValidator()

	// Creating a new admin user to use on the validation tests
    db := initializers.NewDatabaseConn()
    defer db.Close()
    
    var adminUser = models.UserBody{
        Name: "The",
        Surname: "Admin",
        Email: "admin@admin.com",
        Password: "Testando@Teste**",
        Birthday: "1990-10-10",
    }

    admResp, err := userModel.InsertUserInDB(db, adminUser)
    if err != nil {
        log.Fatalf("Error creating adm user in initializers tests setup: %v", err)
    }

	adminId = admResp.ID.String()

    if err := userModel.UpdateUserToAdmById(db, admResp.ID); err != nil {
        log.Fatalf("Error updating user to adm in initializers tests setup: %v", err)
    }
	
	// Run tests
    exitCode := m.Run()

	// Teardown
	if err := tests.Teardown(); err != nil {
		log.Fatalf("Error tearing down tests: %v", err)
	}
	
	os.Exit(exitCode)
}

func Test_structValidation(t *testing.T) {
	type args struct {
		data     interface{}
		validate *validator.Validate
	}
	testCases := []struct {
		name string
		args args
		want []ErrorResponse
	}{
		{
			name: "Testing common and custom validation - Success Case",
			args: args{
				validate: validate,
				data: struct {
					Name     string `json:"name" validate:"required"`
					Email    string `json:"email" validate:"required,email"`
					Password string `json:"password" validate:"required,password"`
					CreatorId string `json:"creatorId" validate:"required,isAdminUuid"`
				}{
					Name:     "John",
					Email:    "john@john.com",
					Password: "Johnjohn123%@",
					CreatorId: adminId,
				},
			},
			want: []ErrorResponse{},
		},
		{
			name: "Testing common and custom validation - Error Case",
			args: args{
				validate: validate,
				data: struct {
					Name     string `json:"name" validate:"required"`
					Surname  string `json:"surname" validate:"omitempty"`
					Email    string `json:"email" validate:"required,email"`
					Password string `json:"password" validate:"required,password"`
					Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
					CreatorId string `json:"creatorId" validate:"required,isAdminUuid"`
				}{
					Email:    "banana",
					Password: "12345",
					Birthday: "23/09/1997",
					CreatorId: "asfasd2",
				},
			},
			want: []ErrorResponse{
				{
					Error:        true,
					FailedField:  "name",
					Tag:          "required",
					ErrorMessage: "The name field is required.",
				},
				{
					Error:        true,
					FailedField:  "email",
					Tag:          "email",
					ErrorMessage: "The email field needs to be a valid email.",
				},
				{
					Error:        true,
					FailedField:  "password",
					Tag:          "password",
					ErrorMessage: "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number.",
				},
				{
					Error:        true,
					FailedField:  "birthday",
					Tag:          "datetime",
					ErrorMessage: "The birthday field needs to follow the YYYY-MM-DD format.",
				},
				{
					Error: true,
					FailedField: "creatorid",
					Tag: "isAdminUuid",
					ErrorMessage: "The creatorId field needs to be a valid uuid that belongs to an admin user.",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := structValidation(testCase.args.validate, testCase.args.data)
			assert.ElementsMatch(t, testCase.want, got, "error lists do not match")
		})
	}
}
