package tests

import (
	"bytes"
	"database/sql"
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

var userResponses []models.UserResponse
var actorResponses []models.ActorResponse
var adminId string

func TestMain(m *testing.M) {
	var err error
	TestDb, err = Setup()
	if err != nil {
		log.Fatalf("Error setting up tests: %v", err)
	}

	os.Setenv("PGDATABASE", TestDb)

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
		{
			Name:     "cccccccc",
			Surname:  "o c",
			Email:    "teste3@teste3.com",
			Password: "testando123@Teste",
			Birthday: "1990-10-10",
		},
	}

	userResponses = InsertMockedUsersInDB(db, usersToBeInsertedInDb)

	// Inserting actors in DB for testing. Reminder that these have relationships to users and can only be created by admins.
	// Creating admin user
	var adminUser = models.UserBody{
		Name:     "The",
		Surname:  "Admin",
		Email:    "admin@admin.com",
		Password: "Testando@Teste**",
		Birthday: "1990-10-10",
	}

	admResp, err := UserModel.InsertUserInDB(db, adminUser)
	if err != nil {
		log.Fatalf("Error creating adm user in initializers tests setup: %v", err)
	}

	adminId = admResp.ID.String()

	if err := UserModel.UpdateUserToAdmById(db, admResp.ID); err != nil {
		log.Fatalf("Error updating user to adm in initializers tests setup: %v", err)
	}

	actorsToBeInsertedInDb := []models.ActorBody{
		{
			Name:      "Actor Name 1",
			Surname:   "Actor Surname 1",
			Birthday:  "2001-10-10",
			CreatorId: adminId,
		},
		{
			Name:      "Actor Name 2",
			Surname:   "Actor Surname 2",
			Birthday:  "2001-10-10",
			CreatorId: adminId,
		},
		{
			Name:      "Actor Name 3",
			Surname:   "Actor Surname 3",
			Birthday:  "2001-10-10",
			CreatorId: adminId,
		},
		{
			Name:      "Actor Name 4",
			Surname:   "Actor Surname 4",
			Birthday:  "2001-10-10",
			CreatorId: adminId,
		},
	}
	actorResponses = InsertMockedActorsInDB(db, actorsToBeInsertedInDb)

	App = fiber.New()

	userController := controllers.User{
		DB:       db,
		Validate: validate,
	}

	sessionController := controllers.Session{
		DB:       db,
		Validate: validate,
	}

	actorController := controllers.Actor{
		DB:       db,
		Validate: validate,
	}

	// Routes - Session
	App.Post("/login", sessionController.HandleLogin)
	App.Post("/logout", sessionController.HandleLogout)

	// Routes - User
	App.Post("/users", userController.CreateUser)
	App.Get("/users", userController.ListAllUsersInDB)
	App.Get("/users/:uuid", userController.GetUser)
	App.Delete("/users/:uuid", userController.DeleteUser)
	App.Patch("/users/:uuid", userController.UpdateUser)

	// Routes - Actor
	App.Post("/actors", actorController.CreateActor)
	App.Get("/actors", actorController.ListAllActorsInDB)
	App.Get("/actors/:uuid", actorController.GetActor)
	App.Delete("/actors/:uuid", actorController.DeleteActor)
	App.Patch("/actors/:uuid", actorController.UpdateActor)

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
				IsAdm:    false,
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
			expectedResponse: GlobalErrorHandlerResp{
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
					IsAdm:    false,
				},
				{
					Name:     "bbbbbb",
					Surname:  "o b",
					Email:    "teste2@teste2.com",
					Birthday: "1990-10-10T00:00:00Z",
					IsAdm:    false,
				},
				{
					Name:     "cccccccc",
					Surname:  "o c",
					Email:    "teste3@teste3.com",
					Birthday: "1990-10-10T00:00:00Z",
					IsAdm:    false,
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
			expectedResponse: GlobalErrorHandlerResp{
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
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Limit needs to be a valid integer",
			},
			responseType: "slice",
			testType:     "global-error",
		}, // Since sort casts every non-valid value to a default valid one, it does not need to be tested, as any error case will fall into the updated_at DESC clause.
		{
			description:      "GET BY ID - Passing an uuid that does not exist in DB - Success Case",
			route:            fmt.Sprintf("/users/%v", userResponses[1].ID),
			method:           "GET",
			expectedCode:     200,
			expectedResponse: userResponses[1],
			responseType:     "struct",
			testType:         "success",
		},
		{
			description:  "GET BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/users/%v", uuid.New()),
			method:       "GET",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
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
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		// Delete requests
		{
			description:      "DELETE BY ID - Passing an uuid that exists in DB - Success Case",
			route:            fmt.Sprintf("/users/%v", userResponses[2].ID),
			method:           "DELETE",
			expectedCode:     204,
			expectedResponse: userResponses[2],
			testType:         "delete",
		},
		{
			description:  "DELETE BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/users/%v", "aushauhsuahsaushuha"),
			method:       "DELETE",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			testType: "global-error",
		},
		// Update requests
		{
			description: "UPDATE - Update user info (all keys) - Success Case",
			route:       fmt.Sprintf("/users/%v", userResponses[3].ID),
			method:      "PATCH",
			data: map[string]interface{}{
				"name":     "New name",
				"surname":  "New surname",
				"password": "Testando123**@",
				"birthday": "1990-10-10",
			},
			expectedCode: 200,
			expectedResponse: models.UserResponse{
				Name:     "New name",
				Surname:  "New surname",
				Email:    "teste3@teste3.com.br",
				Birthday: "1990-10-10T00:00:00Z",
				IsAdm:    false,
			},
			testType: "update",
		},
		{
			description:  "UPDATE - Passing an invalid uuid - Error Case",
			route:        "/users/as9du9u192ejs",
			method:       "PATCH",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
		},
	}

	for _, testCase := range testCases {
		db := initializers.NewDatabaseConn()
		defer db.Close()

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

		resp, err := App.Test(req, -1)
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
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Name, actResp.Name, "Name mismatch")
						assert.Equal(t, expected[i].Surname, actResp.Surname, "Surname mismatch")
						assert.Equal(t, expected[i].Email, actResp.Email, "Email mismatch")
						assert.Equal(t, expected[i].Birthday, actResp.Birthday, "Birthday mismatch")
						assert.Equal(t, expected[i].IsAdm, actResp.IsAdm, "IsAdm mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for _, user := range userResponses {
							if user.Name == actResp.Name && user.Surname == actResp.Surname && user.Email == actResp.Email {
								assert.Equal(t, user.ID, actResp.ID, "ID mismatch")
								assert.Equal(t, user.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
								assert.Equal(t, user.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
								assert.Equal(t, user.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
								assert.Equal(t, user.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
								break
							}
						}
					}
				}

				compareUserResponses(t, testCase.expectedResponse.([]models.UserResponse), respSlice)
			} else {
				compareUserResponses := func(t *testing.T, expected, actual models.UserResponse) {
					expected.ID = uuid.Nil

					assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
					assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
					assert.Equal(t, expected.Email, actual.Email, "Email mismatch")
					assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")
					assert.Equal(t, expected.IsAdm, actual.IsAdm, "Birthday mismatch")
					assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")

					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					for _, user := range userResponses {
						if user.Name == actual.Name && user.Surname == actual.Surname && user.Email == actual.Email {
							assert.Equal(t, user.ID, actual.ID, "ID mismatch")
							assert.Equal(t, user.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
							assert.Equal(t, user.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
							assert.Equal(t, user.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
							assert.Equal(t, user.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
							break
						}
					}
				}

				compareUserResponses(t, testCase.expectedResponse.(models.UserResponse), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}

		if testCase.testType == "delete" {
			userResp, err := UserModel.GetUserByEmail(db, testCase.expectedResponse.(models.UserResponse).Email)
			if err != nil {
				if err == sql.ErrNoRows {
					assert.Fail(t, "User not found in database when getting by id", testCase.expectedResponse.(models.UserResponse).ID)
				}

				assert.Fail(t, "Error when getting user by id", err)
			}

			assert.Equal(t, userResp.DeletedAt.Valid, true, "deletedAt date is not valid after executing delete request on user")
			assert.NotEqual(t, userResp.DeletedAt.Time, time.Time{}, "deleteAt time should be the time of deletion, not a 0 value")
		}

		if testCase.testType == "update" {
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

			compareUserResponses := func(t *testing.T, expected, actual models.UserResponse) {
				expected.ID = uuid.Nil

				assert.Equal(t, expected.Name, actual.Name, "Name should be updated")
				assert.Equal(t, expected.Surname, actual.Surname, "Surname should be updated")
				assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday should be updated")
				assert.Equal(t, expected.IsAdm, actual.IsAdm, "IsAdm should be updated")
				assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")
			}

			compareUserResponses(t, testCase.expectedResponse.(models.UserResponse), respStruct)
		}
	}
}
