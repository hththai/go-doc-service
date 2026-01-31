package test

import (
	"2_Go/api/v1/utils"
	"2_Go/internal/auth"
	"database/sql"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository mocks the auth repository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) Register(tx *sql.Tx, account auth.Account) (*auth.Account, error) {
	args := m.Called(tx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

func (m *MockAuthRepository) ValidateUser(db *sql.DB, username string) (*auth.Account, error) {
	args := m.Called(db, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

func (m *MockAuthRepository) UpdatePassword(tx *sql.Tx, account auth.Account) (*auth.Account, error) {
	args := m.Called(tx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

func (m *MockAuthRepository) UpdatePasswordById(tx *sql.Tx, account auth.Account) (*auth.Account, error) {
	args := m.Called(tx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

func (m *MockAuthRepository) GetUsrPassword(db *sql.DB, username string) (int, string, error) {
	args := m.Called(db, username)
	return args.Int(0), args.String(1), args.Error(2)
}

func (m *MockAuthRepository) GetUsrPasswordById(db *sql.DB, userId int) (string, error) {
	args := m.Called(db, userId)
	return args.String(0), args.Error(1)
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		body          string
		mockRepo      func(repo *MockAuthRepository)
		mockDB        func() (*sql.DB, sqlmock.Sqlmock)
		expectErr     bool
		errorContains string
	}{
		{
			name: "success - valid registration",
			body: `{"username":"testuser","password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("Register", mock.Anything, mock.MatchedBy(func(acc auth.Account) bool {
					return acc.Username == "testuser" && len(acc.Password) > 0
				})).Return(&auth.Account{Username: "testuser"}, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit()
				return db, mock
			},
			expectErr: false,
		},
		{
			name: "error - invalid JSON",
			body: `{"username":"test"`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as it should fail before repository call
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr: true,
		},
		{
			name: "error - missing username",
			body: `{"password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as validation should fail
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "Username",
		},
		{
			name: "error - password too short",
			body: `{"username":"testuser","password":"1234"}`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as validation should fail
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "Password",
		},
		{
			name: "error - missing password",
			body: `{"username":"testuser"}`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as validation should fail
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "Password",
		},
		{
			name: "error - empty request body",
			body: `{}`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as validation should fail
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "Username",
		},
		{
			name: "error - transaction begin fails",
			body: `{"username":"testuser","password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as transaction should fail before repository call
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin().WillReturnError(errors.New("failed to start transaction"))
				return db, mock
			},
			expectErr:     true,
			errorContains: "failed to start transaction",
		},
		{
			name: "error - repository register fails (username exists)",
			body: `{"username":"existinguser","password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("Register", mock.Anything, mock.Anything).
					Return(nil, errors.New("username exists"))
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectRollback()
				return db, mock
			},
			expectErr:     true,
			errorContains: "username exists",
		},
		{
			name: "error - transaction commit fails",
			body: `{"username":"testuser","password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("Register", mock.Anything, mock.Anything).
					Return(&auth.Account{Username: "testuser"}, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
				return db, mock
			},
			expectErr:     true,
			errorContains: "commit failed",
		},
		{
			name: "success - valid registration with email",
			body: `{"username":"testuser","email":"test@example.com","password":"password123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("Register", mock.Anything, mock.MatchedBy(func(acc auth.Account) bool {
					return acc.Username == "testuser" &&
						acc.Email == "test@example.com" &&
						len(acc.Password) > 0
				})).Return(&auth.Account{Username: "testuser", Email: "test@example.com"}, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit()
				return db, mock
			},
			expectErr: false,
		},
		{
			name: "success - minimum valid password length",
			body: `{"username":"testuser","password":"12345"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("Register", mock.Anything, mock.MatchedBy(func(acc auth.Account) bool {
					return acc.Username == "testuser"
				})).Return(&auth.Account{Username: "testuser"}, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit()
				return db, mock
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := &MockAuthRepository{}
			tt.mockRepo(mockRepo)

			// Create auth service with mocked repository
			authService := auth.NewAuthService(mockRepo)

			// Setup mock DB
			db, sqlMock := tt.mockDB()
			defer db.Close()

			// Create Gin context with test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/register", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute the function
			err := utils.Register(c, authService, db)

			// Assertions
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all SQL mock expectations were met
			if err := sqlMock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled sqlmock expectations: %s", err)
			}

			// Verify all repository mock expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
