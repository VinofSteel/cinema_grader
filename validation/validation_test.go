package validation

import (
	"log"
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/tests"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var validate *validator.Validate
var userModel models.UserModel
var actorModel models.ActorModel

var adminId string
var actor1Id string
var actor2Id string

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
		Name:     "The",
		Surname:  "Admin",
		Email:    "admin@admin.com",
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

	// Creating some actors to use on validation tests
	var actor1 = models.ActorBody{
		Name:      "Actor Name 1",
		Surname:   "Actor Surname 1",
		Birthday:  "2001-10-10",
		CreatorId: adminId,
	}
	var actor2 = models.ActorBody{
		Name:      "Actor Name 2",
		Surname:   "Actor Surname 2",
		Birthday:  "2001-10-10",
		CreatorId: adminId,
	}

	actor1Res, err := actorModel.InsertActorInDB(db, actor1)
	if err != nil {
		log.Fatalf("Error creating actor1 in initializers tests setup: %v", err)
	}
	actor1Id = actor1Res.ID.String()

	actor2Res, err := actorModel.InsertActorInDB(db, actor2)
	if err != nil {
		log.Fatalf("Error creating actor2 in initializers tests setup: %v", err)
	}
	actor2Id = actor2Res.ID.String()

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
					Name      string   `json:"name" validate:"required"`
					Email     string   `json:"email" validate:"required,email"`
					Password  string   `json:"password" validate:"required,password"`
					CreatorId string   `json:"creatorId" validate:"required,isadminuuid"`
					Actors    []string `json:"actors" validate:"required,unique,validactorslice"`
					MovieID   string   `json:"movieId" validate:"required,isvaliduuid"`
					Grade     float64  `json:"grade" validate:"omitempty,isvalidgrade"`
				}{
					Name:      "John",
					Email:     "john@john.com",
					Password:  "Johnjohn123%@",
					CreatorId: adminId,
					Actors:    []string{actor1Id, actor2Id},
					MovieID:   uuid.New().String(),
					Grade:     5.0,
				},
			},
			want: []ErrorResponse{},
		},
		{
			name: "Testing common and custom validation - Error Case",
			args: args{
				validate: validate,
				data: struct {
					Name      string   `json:"name" validate:"required"`
					Surname   string   `json:"surname" validate:"omitempty"`
					Email     string   `json:"email" validate:"required,email"`
					Password  string   `json:"password" validate:"required,password"`
					Birthday  string   `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
					CreatorId string   `json:"creatorId" validate:"required,isadminuuid"`
					Actors    []string `json:"actors" validate:"required,unique,validactorslice"`
					MovieID   string   `json:"movieId" validate:"required,isvaliduuid"`
					Grade     float64  `json:"grade" validate:"omitempty,isvalidgrade"`
				}{
					Email:     "banana",
					Password:  "12345",
					Birthday:  "23/09/1997",
					CreatorId: "asfasd2",
					Actors:    []string{"banana", "batata"},
					MovieID:   "asuhduashd",
					Grade:     0.5,
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
					Error:        true,
					FailedField:  "creatorId",
					Tag:          "isadminuuid",
					ErrorMessage: "The creatorId field needs to be a valid uuid that belongs to an admin user.",
				},
				{
					Error:        true,
					FailedField:  "actors",
					Tag:          "validactorslice",
					ErrorMessage: "The actors field needs to be a valid array that contains uuids of existing actors.",
				},
				{
					Error:        true,
					FailedField:  "movieId",
					Tag:          "isvaliduuid",
					ErrorMessage: "The movieId field needs to be a valid uuid.",
				},
				{
					Error:        true,
					FailedField:  "grade",
					Tag:          "isvalidgrade",
					ErrorMessage: "The grade field needs to a float between 1.0 and 5.0, with only one decimal field.",
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
