package initializers

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

var passwordItems = []struct {
	have string
	want bool
}{
	{"123", false},
	{"Testando", false},
	{"@zerty123", false},
	{"Testando123", false},
	{"azerty%123BCA", true},
}

func Test_passwordValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("password", passwordValidation)

	for _, item := range passwordItems {
		err := validate.Var(item.have, "password")

		if item.want && err != nil {
			t.Errorf("Unexpected error: %v for item: %v", err, item)
		} else if !item.want && err == nil {
			t.Errorf("Expected error, but got nil for item: %v", item)
		}
	}
}
