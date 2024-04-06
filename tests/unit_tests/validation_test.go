package unit_tests

import (
	"os"
	"reflect"
	"testing"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/validation"
)

func TestMain(m *testing.M) {
	initializers.InitializeValidator()
	os.Exit(m.Run())
}

func Test_StructValidation(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want []validation.ErrorResponse
	}{
		{
			name: "Success Case - Testing common and custom validation",
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
			want: []validation.ErrorResponse{},
		},
		{
			name: "Invalid Data",
			args: args{
				data: struct{
					Name     string `json:"name" validate:"required"`
					Surname  string `json:"surname" validate:"omitempty"`
					Email    string `json:"email" validate:"required,email"`
					Password string `json:"password" validate:"required,password"`
					Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
				}{
					Email: "banana",
					Password: "12345",
					Birthday: "23/09/1997",
				},
			},
			want: []validation.ErrorResponse{
				{
					Error: true,
					FailedField: "name",
					Tag: "required",
					ErrorMessage: "The name field is required.",
				},
				{
					Error: true,
					FailedField: "email",
					Tag: "email",
					ErrorMessage: "The email field needs to be a valid email.",
				},
				{
					Error: true,
					FailedField: "password",
					Tag: "password",
					ErrorMessage: "The password field needs to have at least 8 characters in length, at least one symbol, one lowercased letter, one uppercased letter and one number.",
				},
				{
					Error: true,
					FailedField: "birthday",
					Tag: "datetime",
					ErrorMessage: "The birthday field needs to follow the YYYY-MM-DD format.",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.StructValidation(tt.args.data)
			for i, err := range tt.want {
				if i >= len(got) {
					t.Fatalf("expected %d errors, got %d", len(tt.want), len(got))
				}
				
				if !reflect.DeepEqual(err, got[i]) {
					t.Errorf("error at index %d: got %#v, want %#v", i, got[i], err)
					t.Errorf("got error message: %q", err.ErrorMessage)
					t.Errorf("want error message: %q", got[i].ErrorMessage)
				}
			}
			
			if len(got) > len(tt.want) {
				t.Fatalf("expected %d errors, got %d", len(tt.want), len(got))
			}
				
		})
	}
}
