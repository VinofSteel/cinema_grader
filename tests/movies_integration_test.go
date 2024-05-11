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
				"actors":      []string{actorResponses[0].ID.String(), actorResponses[1].ID.String()},
			},
			expectedCode: 201,
			expectedResponse: models.MovieResponseWithActors{
				Title:       "Movie 1",
				Director:    "Director 1",
				ReleaseDate: "1990-01-01T00:00:00Z",
				CreatorId:   adminId,
				Actors:      []models.ActorResponse{actorResponses[0], actorResponses[1]},
			},
			testType: "success",
		},
		{
			description: "POST - Movie already exists in DB - Error Case",
			route:       "/movies",
			method:      "POST",
			data: map[string]interface{}{
				"title":       "Movie 1",
				"director":    "Director 1",
				"releaseDate": "1990-01-01",
				"creatorId":   adminId,
				"actors":      []string{actorResponses[0].ID.String(), actorResponses[1].ID.String()},
			},
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Movie with this title already exists",
			},
			testType: "global-error",
		},
		{
			description: "POST WITH ID - Add new actors to a movie - Success Case",
			route:       fmt.Sprintf("/movies/%v/actors", movieResponses[4].ID),
			method:      "POST",
			data: map[string]interface{}{
				"actors": []string{actorResponses[0].ID.String()},
			},
			expectedCode:     204,
			expectedResponse: movieResponses[4],
			testType:         "success-movies-actors",
		},
		{
			description: "POST WITH ID - Add repeat actors to a movie - Error Case",
			route:       fmt.Sprintf("/movies/%v/actors", movieResponses[3].ID),
			method:      "POST",
			data: map[string]interface{}{
				"actors": []string{actorResponses[3].ID.String()},
			},
			expectedCode:     400,
			expectedResponse: movieResponses[3],
			testType:         "success-movies-actors", // This still passes because the test only checks if the relationship exists and, well, it already does.
		},
		{
			description:  "POST WITH ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/movies/%v/actors", uuid.New()),
			method:       "POST",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Movie id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "POST WITH ID - Passing an invalid uuid - Error Case",
			route:        "/movies/testestetsts/actors",
			method:       "POST",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		// Get requests
		{
			description:  "GET - All movies with basic query params and with actors - Success Case", // We don't do gigantic offsets and limits to not need to mock 10 things
			route:        "/movies?offset=0&limit=2&sort=title,asc&deleted=false&with_actors=true",
			method:       "GET",
			expectedCode: 200,
			expectedResponse: []models.MovieResponseWithActors{
				{
					Title:        movieResponses[0].Title,
					Director:     movieResponses[0].Director,
					ReleaseDate:  movieResponses[0].ReleaseDate,
					AverageGrade: movieResponses[0].AverageGrade,
					CreatorId:    adminId,
					Actors:       movieResponses[0].Actors,
				},
				{
					Title:        movieResponses[1].Title,
					Director:     movieResponses[1].Director,
					ReleaseDate:  movieResponses[1].ReleaseDate,
					AverageGrade: movieResponses[1].AverageGrade,
					CreatorId:    adminId,
					Actors:       movieResponses[1].Actors,
				},
			},
			responseType: "slice",
			testType:     "success",
		}, // Not gonna test the movies without actors because it would be a huge hassle for what's basically the same test but without the actors key (which is arguably the hardest part of this)
		{
			description:  "GET - Passing an offset that is not a number - Error Case",
			route:        "/movies?offset=2.254",
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
			route:        "/movies?limit=aushaushaush",
			method:       "GET",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Limit needs to be a valid integer",
			},
			responseType: "slice",
			testType:     "global-error",
		}, // Since sort and with_actors casts every non-valid value to a default valid one, it does not need to be tested, as any error case will fall into the updated_at DESC clause.
		{
			description:      "GET BY ID - Passing an uuid that exists in DB - Success Case",
			route:            fmt.Sprintf("/movies/%v", movieResponses[1].ID),
			method:           "GET",
			expectedCode:     200,
			expectedResponse: movieResponses[1],
			responseType:     "struct",
			testType:         "success",
		},
		{
			description:  "GET BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/movies/%v", uuid.New()),
			method:       "GET",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Movie id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "GET BY ID - Passing an invalid uuid - Error Case",
			route:        "/movies/testestetsts",
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
			route:            fmt.Sprintf("/movies/%v", movieResponses[2].ID),
			method:           "DELETE",
			expectedCode:     204,
			expectedResponse: movieResponses[2],
			testType:         "delete",
		},
		{
			description:  "DELETE BY ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/movies/%v", "testeasdasd"),
			method:       "DELETE",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			testType: "global-error",
		},
		{
			description: "DELETE WITH ID - Delete existing actors to a movie - Success Case",
			route:       fmt.Sprintf("/movies/%v/actors", movieResponses[4].ID),
			method:      "DELETE",
			data: map[string]interface{}{
				"actors": []string{actorResponses[0].ID.String()},
			},
			expectedCode:     204,
			expectedResponse: movieResponses[4],
			testType:         "success-delete-movies-actors",
		},
		{
			description: "DELETE WITH ID - Deleting actors that already don't exist on a movie - Error Case",
			route:       fmt.Sprintf("/movies/%v/actors", movieResponses[3].ID),
			method:      "DELETE",
			data: map[string]interface{}{
				"actors": []string{actorResponses[2].ID.String()},
			},
			expectedCode:     400,
			expectedResponse: movieResponses[3],
			testType:         "success-delete-movies-actors", // This still passes because the test only checks if the relationship exists and, well, it already does.
		},
		{
			description:  "DELETE WITH ID - Passing an uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/movies/%v/actors", uuid.New()),
			method:       "DELETE",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Movie id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "DELETE WITH ID - Passing an invalid uuid - Error Case",
			route:        "/movies/testestetsts/actors",
			method:       "DELETE",
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid uuid parameter",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		// Update requests
		{
			description: "UPDATE - Update movie info (all keys) - Success Case",
			route:       fmt.Sprintf("/movies/%v", movieResponses[3].ID),
			method:      "PATCH",
			data: map[string]interface{}{
				"title":       "New title",
				"director":    "New director",
				"releaseDate": "1990-10-10",
			},
			expectedCode: 200,
			expectedResponse: models.MovieResponse{
				Title:       "New title",
				Director:    "New director",
				ReleaseDate: "1990-10-10T00:00:00Z",
			},
			testType: "update",
		},
		{
			description:  "UPDATE - Passing an invalid uuid - Error Case",
			route:        "/movies/09ehrgf",
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
		if testCase.data != nil {
			var err error
			jsonData, err = json.Marshal(testCase.data)
			if err != nil {
				t.Fatalf("Error marshalling JSON data: %v", err)
			}
		}

		var req *http.Request
		if testCase.data != nil {
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
			var respStruct models.MovieResponseWithActors
			var respSlice []models.MovieResponseWithActors

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
				compareMovieResponses := func(t *testing.T, expected, actual []models.MovieResponseWithActors) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Title, actResp.Title, "Title mismatch")
						assert.Equal(t, expected[i].Director, actResp.Director, "Director mismatch")
						assert.Equal(t, expected[i].ReleaseDate, actResp.ReleaseDate, "ReleaseDate mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						for j, actor := range actResp.Actors {
							assert.Equal(t, expected[i].Actors[j].ID, actor.ID, "Actor ID mismatch")
							assert.Equal(t, expected[i].Actors[j].Name, actor.Name, "Actor Name mismatch")
							assert.Equal(t, expected[i].Actors[j].Surname, actor.Surname, "Actor Surname mismatch")
							assert.Equal(t, expected[i].Actors[j].Birthday, actor.Birthday, "Actor Birthday mismatch")
							assert.Equal(t, expected[i].Actors[j].CreatorId, actor.CreatorId, "Actor CreatorId mismatch")
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

				compareMovieResponses(t, testCase.expectedResponse.([]models.MovieResponseWithActors), respSlice)
			} else {
				compareMovieResponses := func(t *testing.T, expected, actual models.MovieResponseWithActors) {
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

				compareMovieResponses(t, testCase.expectedResponse.(models.MovieResponseWithActors), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}

		if testCase.testType == "delete" {
			movieResp, err := MovieModel.GetMovieByIdWithActors(db, testCase.expectedResponse.(models.MovieResponseWithActors).ID)
			if err != nil {
				if err == sql.ErrNoRows {
					assert.Fail(t, "movie not found in database when getting by id", testCase.expectedResponse.(models.MovieResponseWithActors).ID)
				}

				assert.Fail(t, "Error when getting movie by id", err)
			}

			assert.Equal(t, movieResp.DeletedAt.Valid, true, "deletedAt date is not valid after executing delete request on actor")
			assert.NotEqual(t, movieResp.DeletedAt.Time, time.Time{}, "deleteAt time should be the time of deletion, not a 0 value")
		}

		if testCase.testType == "update" {
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

			compareMovieResponses := func(t *testing.T, expected, actual models.MovieResponse) {
				expected.ID = uuid.Nil

				assert.Equal(t, expected.Title, actual.Title, "Title should be updated")
				assert.Equal(t, expected.Director, actual.Director, "Director should be updated")
				assert.Equal(t, expected.ReleaseDate, actual.ReleaseDate, "ReleaseDate should be updated")
				assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should be nil")
			}

			compareMovieResponses(t, testCase.expectedResponse.(models.MovieResponse), respStruct)
		}

		if testCase.testType == "success-movies-actors" {
			movieResp, err := MovieModel.GetMovieByIdWithActors(db, testCase.expectedResponse.(models.MovieResponseWithActors).ID)
			if err != nil {
				if err == sql.ErrNoRows {
					assert.Fail(t, "movie not found in database when getting by id", testCase.expectedResponse.(models.MovieResponseWithActors).ID)
				}

				assert.Fail(t, "Error when getting movie by id", err)
			}

			var actorWasInsertedInMovie bool
			for _, actor := range movieResp.Actors {
				if actor.ID.String() == testCase.data["actors"].([]string)[0] {
					actorWasInsertedInMovie = true
				}
			}

			assert.Equal(t, actorWasInsertedInMovie, true, "actor was not inserted into pivot table and relationship was not formed")
		}

		if testCase.testType == "success-delete-movies-actors" {
			movieResp, err := MovieModel.GetMovieByIdWithActors(db, testCase.expectedResponse.(models.MovieResponseWithActors).ID)
			if err != nil {
				if err == sql.ErrNoRows {
					assert.Fail(t, "movie not found in database when getting by id", testCase.expectedResponse.(models.MovieResponseWithActors).ID)
				}

				assert.Fail(t, "Error when getting movie by id", err)
			}

			var actorWasNotDeletedFromMovie bool
			for _, actor := range movieResp.Actors {
				if actor.ID.String() == testCase.data["actors"].([]string)[0] {
					actorWasNotDeletedFromMovie = true
				}
			}

			assert.Equal(t, actorWasNotDeletedFromMovie, false, "actor was not inserted into pivot table and relationship was not formed")
		}
	}
}
