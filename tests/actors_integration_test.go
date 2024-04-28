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
				"name":     "Mark",
				"surname":  "Whalberg",
				"birthday": "1971-06-05",
				"creatorId": adminId,
			},
			expectedCode: 201,
			expectedResponse: models.ActorResponse{
				Name:     "Mark",
				Surname:  "Whalberg",
				Birthday: "1971-06-05T00:00:00Z",
				CreatorId: adminId,
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

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}
	}
}