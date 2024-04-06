package validation

import (
	"os"
	"reflect"
	"testing"

	"github.com/VinOfSteel/cinemagrader/initializers"
)

func TestMain(m *testing.M) {
	initializers.InitializeValidator()
	os.Exit(m.Run())
}

func Test_structValidation(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want []ErrorResponse
	}{
		{
			name: "SuccessCase Testing-common-and-custom-validation",
			args: args{
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
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := structValidation(testCase.args.data)
			for i, err := range testCase.want {
				funcResponse := got[i]

				if i >= len(got) {
					t.Fatalf("expected %d errors, got %d\n", len(testCase.want), len(got))
				}

				if !reflect.DeepEqual(err, funcResponse) {
					t.Errorf("error at index %d: got %#v, want %#v\n", i, funcResponse, err)
					t.Errorf("got error message: %q\n", err.ErrorMessage)
					t.Errorf("want error message: %q\n", funcResponse.ErrorMessage)
				}
			}

			if len(got) > len(testCase.want) {
				t.Fatalf("expected %d errors, got %d", len(testCase.want), len(got))
			}

		})
	}
}
