package validation

import (
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var validate *validator.Validate

func TestMain(m *testing.M) {
	validate = initializers.NewValidator()
	os.Exit(m.Run())
}

func Test_structValidation(t *testing.T) {
	type args struct {
		data     interface{}
		validate *validator.Validate
	}
	testCases := []struct {
		name string
		args args
		want []ErrorResponse
	}{
		{
			name: "SuccessCase Testing-common-and-custom-validation",
			args: args{
				validate: validate,
				data: struct {
					Name     string `json:"name" validate:"required"`
					Email    string `json:"email" validate:"required,email"`
					Password string `json:"password" validate:"required,password"`
				}{
					Name:     "John",
					Email:    "john@john.com",
					Password: "Johnjohn123%@",
				},
			},
			want: []ErrorResponse{},
		},
		{
			name: "InvalidCase Testing-common-and-custom-validation",
			args: args{
				validate: validate,
				data: struct {
					Name     string `json:"name" validate:"required"`
					Surname  string `json:"surname" validate:"omitempty"`
					Email    string `json:"email" validate:"required,email"`
					Password string `json:"password" validate:"required,password"`
					Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
				}{
					Email:    "banana",
					Password: "12345",
					Birthday: "23/09/1997",
				},
			},
			want: []ErrorResponse{
				{
					Error:        true,
					FailedField:  "name",
					Tag:          "required",
					ErrorMessage: "The name field is required.",
				},
				{
					Error:        true,
					FailedField:  "email",
					Tag:          "email",
					ErrorMessage: "The email field needs to be a valid email.",
				},
				{
					Error:        true,
					FailedField:  "password",
					Tag:          "password",
					ErrorMessage: "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number.",
				},
				{
					Error:        true,
					FailedField:  "birthday",
					Tag:          "datetime",
					ErrorMessage: "The birthday field needs to follow the YYYY-MM-DD format.",
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := structValidation(testCase.args.validate, testCase.args.data)
			assert.ElementsMatch(t, testCase.want, got, "error lists do not match")
		})
	}
}
