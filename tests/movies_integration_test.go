package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

func Test_MoviesRoutes(t *testing.T) {
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
			description: "POST - Create a new movie route - Success Case",
			route:       "/movies",
			method:      "POST",
			data: map[string]interface{}{
				"title":       "Movie 1",
				"director":    "Director 1",
				"releaseDate": "1990-01-01",
				"creatorId":   adminId,
				"actors":      []uuid.UUID{actorResponses[0].ID, actorResponses[1].ID},
			},
			expectedCode: 201,
			expectedResponse: models.MovieResponse{
				Title:       "Movie 1",
				Director:    "Director 1",
				ReleaseDate: "1990-01-01T00:00:00Z",
				CreatorId:   adminId,
				Actors:      []models.ActorResponse{actorResponses[0], actorResponses[1]},
			},
			testType: "success",
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
			var respStruct models.MovieResponse
			var respSlice []models.MovieResponse

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
				compareMovieResponses := func(t *testing.T, expected, actual []models.MovieResponse) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Title, actResp.Title, "Title mismatch")
						assert.Equal(t, expected[i].Director, actResp.Director, "Director mismatch")
						assert.Equal(t, expected[i].ReleaseDate, actResp.ReleaseDate, "ReleaseDate mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for i, actor := range actResp.Actors {
							assert.Equal(t, expected[i].Actors[i].Name, actor.Name, "Actor Name mismatch")
							assert.Equal(t, expected[i].Actors[i].Surname, actor.Surname, "Actor Surname mismatch")
							assert.Equal(t, expected[i].Actors[i].Birthday, actor.Birthday, "Actor Birthday mismatch")
							assert.Equal(t, expected[i].Actors[i].CreatorId, actor.CreatorId, "Actor CreatorId mismatch")
						}

						for _, movie := range movieResponses {
							if movie.Title == actResp.Title && movie.Director == actResp.Director {
								assert.Equal(t, movie.ID, actResp.ID, "ID mismatch")
								assert.Equal(t, movie.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
								assert.Equal(t, movie.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
								assert.Equal(t, movie.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
								assert.Equal(t, movie.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
								break
							}
						}
					}
				}

				compareMovieResponses(t, testCase.expectedResponse.([]models.MovieResponse), respSlice)
			} else {
				compareMovieResponses := func(t *testing.T, expected, actual models.MovieResponse) {
					expected.ID = uuid.Nil

					assert.Equal(t, expected.Title, actual.Title, "Title mismatch")
					assert.Equal(t, expected.Director, actual.Director, "Director mismatch")
					assert.Equal(t, expected.ReleaseDate, actual.ReleaseDate, "ReleaseDate mismatch")
					assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")

					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					for i, actor := range actual.Actors {
						assert.Equal(t, expected.Actors[i].Name, actor.Name, "Actor Name mismatch")
						assert.Equal(t, expected.Actors[i].Surname, actor.Surname, "Actor Surname mismatch")
						assert.Equal(t, expected.Actors[i].Birthday, actor.Birthday, "Actor Birthday mismatch")
						assert.Equal(t, expected.Actors[i].CreatorId, actor.CreatorId, "Actor CreatorId mismatch")
					}

					for _, movie := range movieResponses {
						if movie.Title == actual.Title && movie.Director == actual.Director {
							assert.Equal(t, movie.ID, actual.ID, "ID mismatch")
							assert.Equal(t, movie.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
							assert.Equal(t, movie.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
							assert.Equal(t, movie.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
							assert.Equal(t, movie.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
							break
						}
					}
				}

				compareMovieResponses(t, testCase.expectedResponse.(models.MovieResponse), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}
	}
}
