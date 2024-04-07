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

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/gofiber/fiber/v2"
)

var app *fiber.App

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

func Test_UsersRoutes(t *testing.T) {
	testCases := []struct {
		description      string
		route            string
		method           string
		data             map[string]interface{}
		expectedCode     int
		expectedResponse interface{}
	}{
		{
			description: "POST - Create a new user route",
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
		if resp.StatusCode != testCase.expectedCode {
			t.Errorf("got status code: %v", resp.StatusCode)
			t.Errorf("want status code: %v", testCase.expectedCode)
		}

		// Unmarshallhing the responseBody into an actual struct
		var respStruct models.UserResponse
		if err := json.Unmarshal(responseBody, &respStruct); err != nil {
			t.Fatalf("Error unmarshalling response body: %v", err)
		}

		if testCase.expectedResponse.(models.UserResponse).Name != respStruct.Name {
			t.Errorf("API response is different from expected response. Response: %v, Expected: %v", testCase.expectedResponse.(models.UserResponse).Name, respStruct.Name)
		}

		if testCase.expectedResponse.(models.UserResponse).Surname != respStruct.Surname {
			t.Errorf("API response is different from expected response. Response: %v, Expected: %v", testCase.expectedResponse.(models.UserResponse).Surname, respStruct.Surname)
		}

		if testCase.expectedResponse.(models.UserResponse).Email != respStruct.Email {
			t.Errorf("API response is different from expected response. Response: %v, Expected: %v", testCase.expectedResponse.(models.UserResponse).Email, respStruct.Email)
		}

		if testCase.expectedResponse.(models.UserResponse).Birthday != respStruct.Birthday {
			t.Errorf("API response is different from expected response. Response: %v, Expected: %v", testCase.expectedResponse.(models.UserResponse).Birthday, respStruct.Birthday)
		}
	}
}
