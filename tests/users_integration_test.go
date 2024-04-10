package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
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

	var wg sync.WaitGroup
	for _, user := range usersToBeInsertedInDb {
		wg.Add(1)
		go func() { // This code is ass. Should make an insert function to insert multiple users in DB using SQL, but...
			_, err = userModel.InsertUserInDB(db, user)
			if err != nil {
				log.Fatalf("Error inserting mocked user with email %v in Db: %v", user.Email, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	app = fiber.New()

	userController := controllers.User{
		DB:       db,
		Validate: validate,
	}

	app.Post("/users", userController.CreateUser)
	app.Get("/users", userController.GetAllUsers)

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
						expected[i].ID = uuid.Nil           // Ignore ID
						expected[i].CreatedAt = time.Time{} // Ignore CreatedAt
						expected[i].UpdatedAt = time.Time{} // Ignore UpdatedAt

						assert.Equal(t, expected[i].Name, actResp.Name, "Name mismatch")
						assert.Equal(t, expected[i].Surname, actResp.Surname, "Surname mismatch")
						assert.Equal(t, expected[i].Email, actResp.Email, "Email mismatch")
						assert.Equal(t, expected[i].Birthday, actResp.Birthday, "Birthday mismatch")
					}
				}

				compareUserResponses(t, testCase.expectedResponse.([]models.UserResponse), respSlice)
			} else {
				compareUserResponses := func(t *testing.T, expected, actual models.UserResponse) {
					expected.ID = uuid.Nil           // Ignore ID
					expected.CreatedAt = time.Time{} // Ignore CreatedAt
					expected.UpdatedAt = time.Time{} // Ignore UpdatedAt

					assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
					assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
					assert.Equal(t, expected.Email, actual.Email, "Email mismatch")
					assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")
				}

				compareUserResponses(t, testCase.expectedResponse.(models.UserResponse), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(globalErrorHandlerResp).Message, string(responseBody))
		}

	}
}
