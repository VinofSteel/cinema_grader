package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var app *fiber.App
var userModel models.UserModel
var userResponses []models.UserResponse

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

	// God, forgive me for what I'm about to do.
	// Inserting mocked users in DB for test
	usersToBeInsertedInDb := []models.UserBody{
		{
			Name:     "Duplicate",
			Surname:  "User",
			Email:    "teste@teste.com",
			Password: "testando123@Teste",
			Birthday: "1990-10-10",
		},
		{
			Name:     "aaaaaa",
			Surname:  "o a",
			Email:    "teste1@teste1.com",
			Password: "testando123@Teste",
			Birthday: "1990-10-10",
		},
		{
			Name:     "bbbbbb",
			Surname:  "o b",
			Email:    "teste2@teste2.com",
			Password: "testando123@Teste",
			Birthday: "1990-10-10",
		},
	}

	// Creating a channel to store all the responses of the insertion in the DB so I can check user uuids in tests and pass them on.
	userResponseChannel := make(chan models.UserResponse, len(usersToBeInsertedInDb))
	InsertMockedUsersInDb(db, usersToBeInsertedInDb, &userResponseChannel)

	for user := range userResponseChannel {
		userResponses = append(userResponses, user)
	}

	app = fiber.New()

	userController := controllers.User{
		DB:       db,
		Validate: validate,
	}

	app.Post("/users", userController.CreateUser)
	app.Get("/users", userController.GetAllUsers)
	app.Get("/users/:uuid", userController.GetUserById)

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
		responseType     string
		testType         string
	}{
		// Post requests
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
			expectedCode: 201,
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
		// Get requests
		{
			description:  "GET - All users with basic query params - Success Case", // We don't do gigantic offsets and limits to not need to mock 10 things
			route:        "/users?offset=1&limit=3&sort=name,asc",
			method:       "GET",
			expectedCode: 200,
			expectedResponse: []models.UserResponse{
				{
					Name:     "Astolfo", // Since this user is created in the POST request, commenting the other tests will net this one a failure. Too bad!
					Surname:  "O inho",
					Email:    "astolfinho@astolfinho.com.br",
					Birthday: "1990-10-10T00:00:00Z",
				},
				{
					Name:     "bbbbbb",
					Surname:  "o b",
					Email:    "teste2@teste2.com",
					Birthday: "1990-10-10T00:00:00Z",
				},
				{
					Name:     "Duplicate",
					Surname:  "User",
					Email:    "teste@teste.com",
					Birthday: "1990-10-10T00:00:00Z",
				},
			},
			responseType: "slice",
			testType:     "success",
		},
		{
			description:  "GET - Passing an offset that is not a number - Error Case",
			route:        "/users?offset=2.254",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: globalErrorHandlerResp{
				Message: "Offset needs to be a valid integer",
			},
			responseType: "slice",
			testType:     "global-error",
		},
		{
			description:  "GET - Passing a limit that is not a number - Error Case",
			route:        "/users?limit=aushaushaush",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: globalErrorHandlerResp{
				Message: "Limit needs to be a valid integer",
			},
			responseType: "slice",
			testType:     "global-error",
		}, // Since sort casts every non-valid value to a default valid one, it does not need to be tested, as any error case will fall into the updated_at DESC clause.
		{
			description:      "GET BY ID - Passing a uuid that does not exist in DB - Success Case",
			route:            fmt.Sprintf("/users/%v", userResponses[1].ID),
			method:           "GET",
			expectedCode:     200,
			expectedResponse: userResponses[1],
			responseType:     "struct",
			testType:         "success",
		},
		{
			description:  "GET BY ID - Passing a uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/users/%v", uuid.New()),
			method:       "GET",
			expectedCode: 404,
			expectedResponse: globalErrorHandlerResp{
				Message: "User id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "GET BY ID - Passing an invalid uuid - Error Case",
			route:        "/users/as9du9u192ejs",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: globalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
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
			// Unmarshalling the responseBody into an actual struct
			var respStruct models.UserResponse
			var respSlice []models.UserResponse

			if testCase.responseType == "slice" {
				if err := json.Unmarshal(responseBody, &respSlice); err != nil {
					t.Fatalf("Error unmarshalling response body: %v", err)
				}
			} else {
				if err := json.Unmarshal(responseBody, &respStruct); err != nil {
					t.Fatalf("Error unmarshalling response body: %v", err)
				}
			}

			if testCase.responseType == "slice" {
				compareUserResponses := func(t *testing.T, expected, actual []models.UserResponse) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil // Ignore ID

						// Asserting other fields
						assert.Equal(t, expected[i].Name, actResp.Name, "Name mismatch")
						assert.Equal(t, expected[i].Surname, actResp.Surname, "Surname mismatch")
						assert.Equal(t, expected[i].Email, actResp.Email, "Email mismatch")
						assert.Equal(t, expected[i].Birthday, actResp.Birthday, "Birthday mismatch")

						// Asserting ID, createdAt, and updatedAt
						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for _, user := range userResponses {
							if user.Name == actResp.Name && user.Surname == actResp.Surname && user.Email == actResp.Email {
								// Asserting ID, createdAt, and updatedAt
								assert.Equal(t, user.ID, actResp.ID, "ID mismatch")
								assert.Equal(t, user.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
								assert.Equal(t, user.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
								break
							}
						}
					}
				}

				compareUserResponses(t, testCase.expectedResponse.([]models.UserResponse), respSlice)
			} else {
				compareUserResponses := func(t *testing.T, expected, actual models.UserResponse) {
					expected.ID = uuid.Nil // Ignore ID

					// Asserting other fields
					assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
					assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
					assert.Equal(t, expected.Email, actual.Email, "Email mismatch")
					assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")

					// Asserting ID, createdAt, and updatedAt
					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					for _, user := range userResponses {
						if user.Name == actual.Name && user.Surname == actual.Surname && user.Email == actual.Email {
							// Asserting ID, createdAt, and updatedAt
							assert.Equal(t, user.ID, actual.ID, "ID mismatch")
							assert.Equal(t, user.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
							assert.Equal(t, user.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
							break
						}
					}
				}

				compareUserResponses(t, testCase.expectedResponse.(models.UserResponse), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(globalErrorHandlerResp).Message, string(responseBody))
		}

	}
}
