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

func Test_CommentsRoutes(t *testing.T) {
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
			description: "POST BY ID - Create a new actor route - Success Case",
			route:       fmt.Sprintf("/comments/%v", userResponses[1].ID.String()),
			method:      "POST",
			data: map[string]interface{}{
				"comment": "i8fhdas8ifdhas0i fhasoif hasoif hasiof hasipodf hpaisd hpas dpoa",
				"grade": 4,
				"movieId": movieResponses[1].ID.String(),
			},
			expectedCode: 201,
			expectedResponse: models.CommentResponse{
				Comment: "i8fhdas8ifdhas0i fhasoif hasoif hasiof hasipodf hpaisd hpas dpoa",
				Grade: 4,
				MovieId: movieResponses[1].ID.String(),
				UserId: userResponses[1].ID.String(),
			},
			testType: "success",
		},
		{
			description:  "POST WITH ID - Passing an user uuid that does not exist in DB - Error Case",
			route:        fmt.Sprintf("/comments/%v", uuid.New()),
			method:       "POST",
			expectedCode: 404,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "User id not found in database",
			},
			responseType: "struct",
			testType:     "global-error",
		},
		{
			description:  "POST WITH ID - Passing an invalid user uuid - Error Case",
			route:        "/comments/testestetsts",
			method:       "POST",
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
			var respStruct models.CommentResponse
			var respSlice []models.CommentResponse

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
				compareCommentResponses := func(t *testing.T, expected, actual []models.CommentResponse) {
					for i, actResp := range actual {
						expected[i].ID = uuid.Nil

						assert.Equal(t, expected[i].Comment, actResp.Comment, "Comment mismatch")
						assert.Equal(t, expected[i].Grade, actResp.Grade, "Grade mismatch")
						assert.Equal(t, sql.NullTime{}, actResp.DeletedAt, "DeletedAt should not be nil")
						assert.Equal(t, expected[i].UserId, actResp.UserId, "UserId mismatch")
						assert.Equal(t, expected[i].MovieId, actResp.MovieId, "MovieId mismatch")

						assert.NotEqual(t, uuid.Nil, actResp.ID, "ID should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.CreatedAt, "CreatedAt should not be nil")
						assert.NotEqual(t, time.Time{}, actResp.UpdatedAt, "UpdatedAt should not be nil")

						// for _, actor := range actorResponses {
						// 	if actor.Name == actResp.Name && actor.Surname == actResp.Surname {
						// 		assert.Equal(t, actor.ID, actResp.ID, "ID mismatch")
						// 		assert.Equal(t, actor.CreatedAt.UTC(), actResp.CreatedAt, "CreatedAt mismatch")
						// 		assert.Equal(t, actor.UpdatedAt.UTC(), actResp.UpdatedAt, "UpdatedAt mismatch")
						// 		assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
						// 		assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
						// 		break
						// 	}
						// }
					}
				}

				compareCommentResponses(t, testCase.expectedResponse.([]models.CommentResponse), respSlice)
			} else {
				compareCommentResponses := func(t *testing.T, expected, actual models.CommentResponse) {
					expected.ID = uuid.Nil

					assert.Equal(t, expected.Comment, actual.Comment, "Comment mismatch")
					assert.Equal(t, expected.Grade, actual.Grade, "Grade mismatch")
					assert.Equal(t, sql.NullTime{}, actual.DeletedAt, "DeletedAt should not be nil")
					assert.Equal(t, expected.UserId, actual.UserId, "UserId mismatch")
					assert.Equal(t, expected.MovieId, actual.MovieId, "MovieId mismatch")

					assert.NotEqual(t, uuid.Nil, actual.ID, "ID should not be nil")
					assert.NotEqual(t, time.Time{}, actual.CreatedAt, "CreatedAt should not be nil")
					assert.NotEqual(t, time.Time{}, actual.UpdatedAt, "UpdatedAt should not be nil")

					// for _, actor := range actorResponses {
					// 	if actor.Name == actual.Name && actor.Surname == actual.Surname {
					// 		assert.Equal(t, actor.ID, actual.ID, "ID mismatch")
					// 		assert.Equal(t, actor.CreatedAt.UTC(), actual.CreatedAt, "CreatedAt mismatch")
					// 		assert.Equal(t, actor.UpdatedAt.UTC(), actual.UpdatedAt, "UpdatedAt mismatch")
					// 		assert.Equal(t, actor.DeletedAt.Valid, false, "DeletedAt should not be a valid date")
					// 		assert.Equal(t, actor.DeletedAt.Time, time.Time{}, "DeletedAt should be a 0 value")
					// 		break
					// 	}
					// }
				}

				compareCommentResponses(t, testCase.expectedResponse.(models.CommentResponse), respStruct)
			}
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}
	}
}