package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

var app *fiber.App
var userModel models.UserModel

type globalErrorHandlerResp struct {
	Message string `json:"message"`
}

func TestMain(m *testing.M) {
	var err error
	testDb, err = Setup()
	if err != nil {
		log.Fatalf("Error setting up tests: %v", err)
	}

	os.Setenv("PGDATABASE", testDb)

	validate := initializers.NewValidator()
	db := initializers.NewDatabaseConn()
	defer db.Close()

	// Inserting mocked user in DB for test
	mockedUser := models.UserBody{
		Name:     "Duplicate",
		Surname:  "User",
		Email:    "teste@teste.com",
		Password: "testando123@Teste",
		Birthday: "1990-10-10",
	}

	_, err = userModel.InsertUserInDB(db, mockedUser)
	if err != nil {
		log.Fatalf("Error inserting mocked users in Db: %v", err)
	}

	app = fiber.New()

	userController := controllers.User{
		DB:       db,
		Validate: validate,
	}

	app.Post("/users", userController.CreateUser)

	// Run tests
	exitCode := m.Run()

	// Teardown
	if err := Teardown(); err != nil {
		log.Fatalf("Error tearing down tests: %v", err)
	}

	os.Exit(exitCode)
}

// No need to test for validation errors because the validation function is already unit tested elsewhere
func Test_UsersRoutes(t *testing.T) {
	testCases := []struct {
		description      string
		route            string
		method           string
		data             map[string]interface{}
		expectedCode     int
		expectedResponse interface{}
		testType         string
	}{
		{
			description: "POST - Create a new user route - Success Case",
			route:       "/users",
			method:      "POST",
			data: map[string]interface{}{
				"name":     "Astolfo",
				"surname":  "O inho",
				"email":    "astolfinho@astolfinho.com.br",
				"password": "Astolfinho123@*",
				"birthday": "1990-10-10",
			},
			expectedCode: 200,
			expectedResponse: models.UserResponse{
				Name:     "Astolfo",
				Surname:  "O inho",
				Email:    "astolfinho@astolfinho.com.br",
				Birthday: "1990-10-10T00:00:00Z",
			},
			testType: "success",
		},
		{
			description: "POST - User already exists in DB - Error Case",
			route:       "/users",
			method:      "POST",
			data: map[string]interface{}{
				"name":     "Duplicate",
				"surname":  "User",
				"email":    "teste@teste.com",
				"password": "testando123@Teste",
				"birthday": "1999-10-10",
			},
			expectedCode: 400,
			expectedResponse: globalErrorHandlerResp{
				Message: "User with this email already exists",
			},
			testType: "global-error",
		},
	}

	for _, testCase := range testCases {
		var jsonData []byte
		if testCase.method != "GET" && testCase.method != "DELETE" {
			var err error
			jsonData, err = json.Marshal(testCase.data)
			if err != nil {
				t.Fatalf("Error marshalling JSON data: %v", err)
			}
		}

		var req *http.Request
		if testCase.method != "GET" && testCase.method != "DELETE" {
			req = httptest.NewRequest(testCase.method, testCase.route, bytes.NewBuffer(jsonData))
		} else {
			req = httptest.NewRequest(testCase.method, testCase.route, nil)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Error testing app requisition: %v", err)
		}

		// Converting body to a byte slice
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %v", err)
		}

		// Verifying status code
		assert.Equal(t, testCase.expectedCode, resp.StatusCode, "status code")

		if testCase.testType == "success" {
			// Unmarshallhing the responseBody into an actual struct
			var respStruct models.UserResponse
			if err := json.Unmarshal(responseBody, &respStruct); err != nil {
				t.Fatalf("Error unmarshalling response body: %v", err)
			}
			compareUserResponses := func(t *testing.T, expected, actual models.UserResponse) {
				expected.ID = ""                 // Ignore ID
				expected.CreatedAt = time.Time{} // Ignore CreatedAt
				expected.UpdatedAt = time.Time{} // Ignore UpdatedAt

				assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
				assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
				assert.Equal(t, expected.Email, actual.Email, "Email mismatch")
				assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")
			}

			compareUserResponses(t, testCase.expectedResponse.(models.UserResponse), respStruct)
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(globalErrorHandlerResp).Message, string(responseBody))
		}

	}
}
