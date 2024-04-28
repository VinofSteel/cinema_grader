package initializers

import (
	"log"
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/tests"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type validateTests struct {
	Have string
	Want bool
}

var testDb string
var userModel models.UserModel
var validate *validator.Validate
var adminId string

func TestMain(m *testing.M) {
	// Setup
	var err error
	testDb, err = tests.Setup()
	if err != nil {
		log.Fatalf("Error setting up tests: %v", err)
	}

	os.Setenv("PGDATABASE", testDb)

	// Validator setup
	validate = NewValidator()

	// Creating a new admin user to use on the validation tests
	db := NewDatabaseConn()
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

	// Run tests
	exitCode := m.Run()

	// Teardown
	if err := tests.Teardown(); err != nil {
		log.Fatalf("Error tearing down tests: %v", err)
	}

	os.Exit(exitCode)
}

func Test_NewDatabaseConn(t *testing.T) {
	os.Setenv("PGDATABASE", testDb)

	// Call the function being tested
	db := NewDatabaseConn()
	defer db.Close()

	// Assert that the connection is not nil
	assert.NotNil(t, db, "Expected a non-nil database connection")

	err := db.Ping()
	assert.NoError(t, err, "Error pinging db to make sure it works: %v", err)
}

var passwordItems = []validateTests{
	{"123", false},
	{"Testando", false},
	{"@zerty123", false},
	{"Testando123", false},
	{"azerty%123BCA", true},
}

var adminUuidItems = []validateTests{
	{"a61b6ed8-cd86-4bd9-833b-910b485471c6", false},
	{"banana", false},
	{"12345", false},
	{"", false},
	{"adminId", true},
}

func Test_passwordValidation(t *testing.T) {
	for _, item := range passwordItems {
		err := validate.Var(item.Have, "password")

		if item.Want {
			assert.NoError(t, err, "Unexpected error for item: %v", item)
		} else {
			assert.Error(t, err, "Expected error for item: %v", item)
		}
	}
}

func Test_adminUuidValidation(t *testing.T) {
	for _, item := range adminUuidItems {
		var err error

		if item.Have == "adminId" {
			err = validate.Var(adminId, "isAdminUuid")
		} else {
			err = validate.Var(item.Have, "isAdminUuid")
		}

		if item.Want {
			assert.NoError(t, err, "Unexpected error for item: %v", item)
		} else {
			assert.Error(t, err, "Expected error for item: %v", item)
		}
	}
}
