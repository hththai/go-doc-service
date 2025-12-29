package autho_test

import (
	"2_Go/internal/auth"
	"database/sql"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidate(t *testing.T) {
	accounts := []struct {
		name     string
		account  auth.Account
		expected bool
	}{
		{
			name:    "Invalid username",
			account: auth.Account{Email: "hello"}, expected: false,
		},
		{
			name:     "Empty Input",
			account:  auth.Account{},
			expected: false,
		},
		{
			name:     "Invalid password less than 5 words",
			account:  auth.Account{Username: "hello", Password: "acb"},
			expected: false,
		},
		{
			name:     "Valid Account",
			account:  auth.Account{Username: "hello", Password: "thisis5"},
			expected: true,
		},
	}

	for _, tc := range accounts {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.account.Validate()

			if err != nil && tc.expected {
				t.Errorf("UNEXPECTED error: %v with the expected %v with error: %v", tc.name, tc.expected, err)
			}

			if err == nil && !tc.expected {
				t.Fatalf("EXPECTED error, got nil %v %v and %v with error %v", tc.name, tc.expected, !tc.expected, err)
			}
		})
	}
}

// validate password
func TestValidatePassword(t *testing.T) {
	accounts := []struct {
		name     string
		account  auth.Account
		expected bool
	}{
		{
			name:     "Wrong password case 1",
			account:  auth.Account{Password: "He"},
			expected: false,
		},
		{
			name:     "Valid password",
			account:  auth.Account{Password: "helooworld"},
			expected: true,
		},
	}

	for _, tc := range accounts {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.account.ValidatePassword()

			if err != nil && tc.expected {
				t.Fatalf("UNEXPECTED error with case::: %v with error::: %v", tc.name, err)
			}

			if err == nil && !tc.expected {
				t.Fatalf("EXPECTED error with case::: %v with error::: %v", tc.name, err)
			}
		})
	}
}

type MockAuthoRepo struct {
	RegisterFn       func(*sql.Tx, auth.Account) (*auth.Account, error)
	ValidateUserFn   func(*sql.DB, string) (*auth.Account, error)
	UpdatePasswordFn func(*sql.Tx, auth.Account) (*auth.Account, error)
	GetUsrPasswordFn func(*sql.DB, string) (string, error)
}

func (m *MockAuthoRepo) Register(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
	return m.RegisterFn(tx, a)
}

func (m *MockAuthoRepo) ValidateUser(db *sql.DB, username string) (*auth.Account, error) {
	return m.ValidateUserFn(db, username)
}

func (m *MockAuthoRepo) UpdatePassword(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
	return m.UpdatePasswordFn(tx, a)
}

func (m *MockAuthoRepo) GetUsrPassword(db *sql.DB, username string) (string, error) {
	return m.GetUsrPasswordFn(db, username)
}

func TestHandleHashPassword(t *testing.T) {
	accounts := []struct {
		name     string
		account  auth.Account
		mockRepo func() *MockAuthoRepo
		expected bool
	}{
		{
			name:    "Empty Password",
			account: auth.Account{Username: "hello"},
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{}
			},
			expected: false,
		},
		{
			name:    "Space Password",
			account: auth.Account{Username: "hello", Password: ""},
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{RegisterFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return nil, fmt.Errorf("Invalid account")
				}}
			},
			expected: false,
		},
		{
			name:    "Invalid password less than 5 words",
			account: auth.Account{Username: "hello", Password: "acb"},
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{RegisterFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return nil, fmt.Errorf("Invalid account")
				}}
			},
			expected: false,
		},
		{
			name:    "Valid Account",
			account: auth.Account{Username: "hello", Password: "thisis5"},
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{RegisterFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return &a, nil
				}}
			},
			expected: true,
		},
	}

	for _, tc := range accounts {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.mockRepo()
			svc := auth.NewAuthService(repo)

			result, err := svc.Register(nil, tc.account)

			if err != nil && tc.expected {
				t.Errorf("UNEXPECTED error: %v with the expected %v with error: %v", tc.name, tc.expected, err)
			}

			if err == nil && !tc.expected {
				t.Fatalf("EXPECTED error, got nil %v %v and %v with error %v", tc.name, tc.expected, !tc.expected, err)
			}

			// Checking if password different hash password.
			if err == nil && result != nil {
				if result.Password == tc.account.Password {
					t.Fatalf("expected hashed password, got plain text")
				}
			}
		})
	}
}

// Test change password
func TestChangePassword(t *testing.T) {
	accounts := []struct {
		name        string
		account     auth.Account
		newPassword string
		mockRepo    func() *MockAuthoRepo
		expected    bool
	}{
		{
			name:        "Good Reset Password",
			account:     auth.Account{Username: "username", Password: "This_isSimplePassword"},
			newPassword: "thisd",
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{UpdatePasswordFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return nil, nil
				}}
			},
			expected: true,
		},
		{
			name:        "Invalid New Password",
			account:     auth.Account{Username: "username", Password: "This_isSimplePassword"},
			newPassword: "th",
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{UpdatePasswordFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return nil, nil
				}}
			},
			expected: false,
		},
		{
			name:        "Invalid current account",
			account:     auth.Account{Username: "username"},
			newPassword: "th",
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{}
			},
			expected: false,
		},
	}

	for _, tc := range accounts {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.mockRepo()
			svc := auth.NewAuthService(repo)

			result, err := svc.ChangePasswordService(nil, tc.account, tc.newPassword)

			if err != nil && tc.expected {
				t.Fatalf("unexpected error %v with case %v", err, tc.name)
			}

			if err == nil && !tc.expected {
				t.Fatalf("expected error %v with case %v", err, tc.name)
			}

			if err == nil && result != nil {
				if result.Password == tc.newPassword {
					t.Fatalf("Unexpected error with hash new password")
				}
			}

		})
	}

}

// Test comparing password
func TestCheckPassword(t *testing.T) {
	accounts := []struct {
		name        string
		account     auth.Account
		newPassword string
		mockRepo    func() *MockAuthoRepo
		expected    bool
	}{
		{
			name:        "Correct Password",
			account:     auth.Account{Username: "username", Password: "This_isSimplePassword"},
			newPassword: "This_isSimplePassword",
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{RegisterFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return &a, nil
				}}
			},
			expected: true,
		},
		{
			name:        "Invalid Password",
			account:     auth.Account{Username: "username", Password: "This_isSimplePassword"},
			newPassword: "th",
			mockRepo: func() *MockAuthoRepo {
				return &MockAuthoRepo{RegisterFn: func(tx *sql.Tx, a auth.Account) (*auth.Account, error) {
					return &a, nil
				}}
			},
			expected: false,
		},
	}

	for _, tc := range accounts {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.mockRepo()
			svc := auth.NewAuthService(repo)

			registeredAccount, err := svc.Register(nil, tc.account)

			if err != nil {
				t.Fatalf("error of set up account %v", err)
			}

			err = svc.CheckPassword(registeredAccount, tc.newPassword)

			if err != nil && tc.expected {
				t.Fatalf("unexpected error %v with case %v", err, tc.name)
			}

			if err == nil && !tc.expected {
				t.Fatalf("expected error %v with case %v", err, tc.name)
			}

		})
	}

}

func TestValidateAccountService(t *testing.T) {
	var db *sql.DB = nil

	mockRepo := &MockAuthoRepo{}
	svc := auth.NewAuthService(mockRepo)

	hash := func(pwd string) string {
		h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		return string(h)
	}

	t.Run("user not found", func(t *testing.T) {
		mockRepo.GetUsrPasswordFn = func(db *sql.DB, username string) (string, error) {
			return "", fmt.Errorf("no user")
		}

		err := svc.ValidateAccountService(db, "huy", "12345")
		if err == nil || err.Error() != "invalid account" {
			t.Errorf("expected invalid account error, got %v", err)
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		mockRepo.GetUsrPasswordFn = func(db *sql.DB, username string) (string, error) {
			return hash("correctpwd"), nil
		}

		err := svc.ValidateAccountService(db, "huy", "wrongpwd")
		if err == nil || err.Error() != "incorrect password" {
			t.Errorf("expected incorrect password error, got %v", err)
		}
	})

	t.Run("correct password", func(t *testing.T) {
		mockRepo.GetUsrPasswordFn = func(db *sql.DB, username string) (string, error) {
			return hash("correctpwd"), nil
		}

		err := svc.ValidateAccountService(db, "huy", "correctpwd")
		if err != nil {
			t.Errorf("expected success, got error %v", err)
		}
	})
}
