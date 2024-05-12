package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type LoginBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type LoginResponse struct {
	UserID uuid.UUID `json:"userId"`
}

func Test_SessionsRoutes(t *testing.T) {
	testCases := []struct {
		description      string
		route            string
		method           string
		data             map[string]interface{}
		expectedCode     int
		expectedResponse interface{}
		testType         string
	}{
		// Post requests
		{
			description: "POST - Login with an existing user - Success Case",
			route:       "/login",
			method:      "POST",
			data: map[string]interface{}{
				"email":    "teste1@teste1.com",
				"password": "testando123@Teste",
			},
			expectedCode:     200,
			expectedResponse: LoginResponse{UserID: userResponses[3].ID},
			testType:         "login-success",
		},
		{
			description: "POST - Login with wrong password - Error Case",
			route:       "/login",
			method:      "POST",
			data: map[string]interface{}{
				"email":    "teste1@teste1.com",
				"password": "EuGostode123@@",
			},
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid email/password",
			},
			testType: "global-error",
		},
		{
			description: "POST - Login with non-existant email in DB - Error Case",
			route:       "/login",
			method:      "POST",
			data: map[string]interface{}{
				"email":    "batatinha@tsdasde1.com",
				"password": "EuGostode123@@",
			},
			expectedCode: 400,
			expectedResponse: GlobalErrorHandlerResp{
				Message: "Invalid email/password",
			},
			testType: "global-error",
		},
		{
			description: "POST - Logout with an auth cookie - Success Case",
			route:       "/logout",
			method:      "POST",
			data: map[string]interface{}{
				"email":    "teste1@teste1.com",
				"password": "testando123@Teste",
			},
			expectedCode: 204,
			testType:     "logout-success",
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

		if testCase.testType == "login-success" {
			var respStruct LoginResponse

			if err := json.Unmarshal(responseBody, &respStruct); err != nil {
				t.Fatalf("Error unmarshalling response body: %v", err)
			}

			_, err := uuid.Parse(testCase.expectedResponse.(LoginResponse).UserID.String())
			if err != nil {
				t.Error("Login route must return a valid uuid of the logged in user")
			}

			cookies := resp.Cookies()
			assert.Len(t, cookies, 1, "unexpected number of cookies")
			assert.Equal(t, "Authorization", cookies[0].Name, "unexpected cookie name")
			assert.NotEqual(t, "", cookies[0].Value, "cookie is empty")
		}

		if testCase.testType == "global-error" {
			assert.Equal(t, testCase.expectedResponse.(GlobalErrorHandlerResp).Message, string(responseBody))
		}

		if testCase.testType == "logout-success" {
			cookies := resp.Cookies()
			assert.Len(t, cookies, 1, "unexpected number of cookies")
			assert.Equal(t, "Authorization", cookies[0].Name, "unexpected cookie name")
			assert.Equal(t, "", cookies[0].Value, "authentication cookie value hasn't been cleared")
		}
	}
}
