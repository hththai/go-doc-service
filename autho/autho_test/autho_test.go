package autho_test

import (
	"go_ocr/internal/autho"
	"testing"
)

func TestValidate(t *testing.T) {
	users := []struct {
		name     string
		user     autho.User
		expected bool
	}{
		{name: "Empty user",
			user: autho.User{}, expected: false},
		{name: "missing user name",
			user: autho.User{
				Username: "",
				Email:    "Hello@gmailmg",
			}, expected: false},
		{name: "Missing email",
			user:     autho.User{Username: "heloo"},
			expected: false},
		{name: "Good",
			user:     autho.User{Username: "heloo", Email: "helloworld@gmail.com"},
			expected: true},
	}

	for _, tc := range users {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate()

			if err != nil && tc.expected {
				t.Errorf("Unexpected error: %v with the expected %v with error: %v", tc.name, tc.expected, err)
			}

			if err == nil && !tc.expected {
				t.Fatalf("expected error, got nil %v %v and %v with error %v", tc.name, tc.expected, !tc.expected, err)
			}
		})
	}
}
