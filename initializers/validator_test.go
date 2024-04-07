package initializers

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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

		if item.want {
			assert.NoError(t, err, "Unexpected error for item: %v", item)
		} else {
			assert.Error(t, err, "Expected error for item: %v", item)
		}
	}
}
