package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_ActorRoutes(t *testing.T) {
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
			description: "POST - Create a new actor route - Success Case",
			route:       "/actors",
			method:      "POST",
			data: map[string]interface{}{
				"name":      "Mark",
				"surname":   "Whalberg",
				"birthday":  "1971-06-05",
				"creatorId": adminId,
			},
			expectedCode: 201,
			expectedResponse: models.ActorResponse{
				Name:      "Mark",
				Surname:   "Whalberg",
				Birthday:  "1971-06-05T00:00:00Z",
				CreatorId: adminId,
			},
			testType: "success",
		}, // No error case, because the only possible errors are handled by validators and middleware
		// Get requests
		{
			description:  "GET - All actors with basic query params - Success Case", // We don't do gigantic offsets and limits to not need to mock 10 things
			route:        "/actors?offset=1&limit=3&sort=name,asc",
			method:       "GET",
			expectedCode: 200,
			expectedResponse: []models.ActorResponse{
				{
					Name:      "Actor Name 2",
					Surname:   "Actor Surname 2",
					Birthday:  "2001-10-10T00:00:00Z",
					CreatorId: adminId,
				},
				{
					Name:      "Actor Name 3",
					Surname:   "Actor Surname 3",
					Birthday:  "2001-10-10T00:00:00Z",
					CreatorId: adminId,
				},
				{
					Name:      "Actor Name 4", // Since we have the actor created in the POST request, commenting the other tests will net this one a failure. Too bad!
					Surname:   "Actor Surname 4",
					Birthday:  "2001-10-10T00:00:00Z",
					CreatorId: adminId,
				},
			},
			responseType: "slice",
			testType:     "success",
		},
		{
			description:  "GET - Passing an offset that is not a number - Error Case",
			route:        "/actors?offset=2.254",
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
			route:        "/actors?limit=aushaushaush",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Limit needs to be a valid integer",
			},
			responseType: "slice",
			testType:     "global-error",
		}, // Since sort casts every non-valid value to a default valid one, it does not need to be tested, as any error case will fall into the updated_at DESC clause.
		{
			description:      "GET BY ID - Passing an uuid that exists in DB - Success Case",
			route:            fmt.Sprintf("/actors/%v", actorResponses[1].ID),
			method:           "GET",
			expectedCode:     200,
			expectedResponse: actorResponses[1],
			responseType:     "struct",
			testType:         "success",
		},
		{
			description:  "GET BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/actors/%v", uuid.New()),
			method:       "GET",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Actor id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "GET BY ID - Passing an invalid uuid - Error Case",
			route:        "/actors/testestetsts",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "GET BY ID WITH MOVIES - Passing an uuid that does not exist in DB - Success Case",
			route:        fmt.Sprintf("/actors/%v/movies", actorResponses[1].ID),
			method:       "GET",
			expectedCode: 200,
			expectedResponse: models.ActorResponseWithMovies{
				Name:      actorResponses[1].Name,
				Surname:   actorResponses[1].Surname,
				Birthday:  actorResponses[1].Birthday,
				CreatorId: adminId,
				Movies: []models.MovieResponse{
					{
						Title:       "Inserted Movie 1",
						Director:    "Inserted Director 1",
						ReleaseDate: "1999-01-01T00:00:00Z",
						CreatorId:   adminId,
					},
				},
			},
			responseType: "struct",
			testType:     "success-with-movies",
		},
		{
			description:  "GET BY ID WITH MOVIES - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/actors/%v/movies", uuid.New()),
			method:       "GET",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Actor id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "GET BY ID WITH MOVIES - Passing an invalid uuid - Error Case",
			route:        "/actors/randomshit/movies",
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
			route:            fmt.Sprintf("/actors/%v", actorResponses[2].ID),
			method:           "DELETE",
			expectedCode:     204,
			expectedResponse: actorResponses[2],
			testType:         "delete",
		},
		{
			description:  "DELETE BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/actors/%v", "testeasdasd"),
			method:       "DELETE",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			testType: "global-error",
		},
		// Update requests
		{
			description: "UPDATE - Update actor info (all keys) - Success Case",
			route:       fmt.Sprintf("/actors/%v", actorResponses[3].ID),
			method:      "PATCH",
			data: map[string]interface{}{
				"name":     "New name",
				"surname":  "New surname",
				"birthday": "1990-10-10",
			},
			expectedCode: 200,
			expectedResponse: models.ActorResponse{
				Name:     "New name",
				Surname:  "New surname",
				Birthday: "1990-10-10T00:00:00Z",
			},
			testType: "update",
		},
		{
			description:  "UPDATE - Passing an invalid uuid - Error Case",
			route:        "/actors/12345677",
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
			var respStruct models.ActorResponse
			var respSlice []models.ActorResponse

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
				compareActorResponses := func(t *testing.T, expected, actual []models.ActorResponse) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Name, actResp.Name, "Name mismatch")
						assert.Equal(t, expected[i].Surname, actResp.Surname, "Surname mismatch")
						assert.Equal(t, expected[i].Birthday, actResp.Birthday, "Birthday mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")
						assert.Equal(t, expected[i].CreatorId, actResp.CreatorId, "CreatorId mismatch")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for _, actor := range actorResponses {
							if actor.Name == actResp.Name && actor.Surname == actResp.Surname {
								assert.Equal(t, actor.ID, actResp.ID, "ID mismatch")
								assert.Equal(t, actor.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
								assert.Equal(t, actor.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
								assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
								assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
								break
							}
						}
					}
				}

				compareActorResponses(t, testCase.expectedResponse.([]models.ActorResponse), respSlice)
			} else {
				compareActorResponses := func(t *testing.T, expected, actual models.ActorResponse) {
					expected.ID = uuid.Nil

					assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
					assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
					assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")
					assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")
					assert.Equal(t, expected.CreatorId, actual.CreatorId, "CreatorId mismatch")

					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					for _, actor := range actorResponses {
						if actor.Name == actual.Name && actor.Surname == actual.Surname {
							assert.Equal(t, actor.ID, actual.ID, "ID mismatch")
							assert.Equal(t, actor.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
							assert.Equal(t, actor.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
							assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
							assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
							break
						}
					}
				}

				compareActorResponses(t, testCase.expectedResponse.(models.ActorResponse), respStruct)
			}
		}

		if testCase.testType == "success-with-movies" {
			// Unmarshalling the responseBody into an actual struct
			var respStruct models.ActorResponseWithMovies
			var respSlice []models.ActorResponseWithMovies

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
				compareActorResponses := func(t *testing.T, expected, actual []models.ActorResponseWithMovies) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Name, actResp.Name, "Name mismatch")
						assert.Equal(t, expected[i].Surname, actResp.Surname, "Surname mismatch")
						assert.Equal(t, expected[i].Birthday, actResp.Birthday, "Birthday mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")
						assert.Equal(t, expected[i].CreatorId, actResp.CreatorId, "CreatorId mismatch")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for i, movie := range actResp.Movies {
							assert.Equal(t, expected[i].Movies[i].Title, movie.Title, "Movie Title mismatch")
							assert.Equal(t, expected[i].Movies[i].Director, movie.Director, "Movie Director mismatch")
							assert.Equal(t, expected[i].Movies[i].ReleaseDate, movie.ReleaseDate, "Movie ReleaseDate mismatch")
							assert.Equal(t, expected[i].Movies[i].CreatorId, movie.CreatorId, "Movie CreatorId mismatch")
						}

						for _, actor := range actorResponses {
							if actor.Name == actResp.Name && actor.Surname == actResp.Surname {
								assert.Equal(t, actor.ID, actResp.ID, "ID mismatch")
								assert.Equal(t, actor.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
								assert.Equal(t, actor.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
								assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
								assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
								break
							}
						}
					}
				}

				compareActorResponses(t, testCase.expectedResponse.([]models.ActorResponseWithMovies), respSlice)
			} else {
				compareActorResponses := func(t *testing.T, expected, actual models.ActorResponseWithMovies) {
					expected.ID = uuid.Nil

					assert.Equal(t, expected.Name, actual.Name, "Name mismatch")
					assert.Equal(t, expected.Surname, actual.Surname, "Surname mismatch")
					assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday mismatch")
					assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")
					assert.Equal(t, expected.CreatorId, actual.CreatorId, "CreatorId mismatch")

					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					for i, movie := range actual.Movies {
						assert.Equal(t, expected.Movies[i].Title, movie.Title, "Movie Title mismatch")
						assert.Equal(t, expected.Movies[i].Director, movie.Director, "Movie Director mismatch")
						assert.Equal(t, expected.Movies[i].ReleaseDate, movie.ReleaseDate, "Movie ReleaseDate mismatch")
						assert.Equal(t, expected.Movies[i].CreatorId, movie.CreatorId, "Movie CreatorId mismatch")
					}

					for _, actor := range actorResponses {
						if actor.Name == actual.Name && actor.Surname == actual.Surname {
							assert.Equal(t, actor.ID, actual.ID, "ID mismatch")
							assert.Equal(t, actor.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
							assert.Equal(t, actor.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
							assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
							assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
							break
						}
					}
				}

				compareActorResponses(t, testCase.expectedResponse.(models.ActorResponseWithMovies), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}

		if testCase.testType == "delete" {
			actorResp, err := ActorModel.GetActorById(db, testCase.expectedResponse.(models.ActorResponse).ID)
			if err != nil {
				if err == sql.ErrNoRows {
					assert.Fail(t, "Actor not found in database when getting by id", testCase.expectedResponse.(models.ActorResponse).ID)
				}

				assert.Fail(t, "Error when getting actor by id", err)
			}

			assert.Equal(t, actorResp.DeletedAt.Valid, true, "deletedAt date is not valid after executing delete request on actor")
			assert.NotEqual(t, actorResp.DeletedAt.Time, time.Time{}, "deleteAt time should be the time of deletion, not a 0 value")
		}

		if testCase.testType == "update" {
			var respStruct models.ActorResponse
			var respSlice []models.ActorResponse

			if testCase.responseType == "slice" {
				if err := json.Unmarshal(responseBody, &respSlice); err != nil {
					t.Fatalf("Error unmarshalling response body: %v", err)
				}
			} else {
				if err := json.Unmarshal(responseBody, &respStruct); err != nil {
					t.Fatalf("Error unmarshalling response body: %v", err)
				}
			}

			compareActorResponses := func(t *testing.T, expected, actual models.ActorResponse) {
				expected.ID = uuid.Nil

				assert.Equal(t, expected.Name, actual.Name, "Name should be updated")
				assert.Equal(t, expected.Surname, actual.Surname, "Surname should be updated")
				assert.Equal(t, expected.Birthday, actual.Birthday, "Birthday should be updated")
				assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should be nil")
			}

			compareActorResponses(t, testCase.expectedResponse.(models.ActorResponse), respStruct)
		}
	}
}
