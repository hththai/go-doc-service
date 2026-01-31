package test

import (
	"2_Go/api/v1/utils"
	"2_Go/internal/auth"
	"2_Go/middleware/authen"
	serverutils "2_Go/utils"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// Test Register
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

// Test Change PasswordByID
func TestChangePasswordById(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Bcrypt hash of "oldpassword123"
	validOldPasswordHash := "$2a$10$tNMPSwPNfwwKmL0i7J00beidHBjxX9gmyESatFe5GBlxETjwRhMNm"

	tests := []struct {
		name          string
		body          string
		userId        interface{} // The userId to set in context
		setUserId     bool        // Whether to set userId in context
		mockRepo      func(repo *MockAuthRepository)
		mockDB        func() (*sql.DB, sqlmock.Sqlmock)
		expectErr     bool
		errorContains string
	}{
		{
			name:      "success - valid password change",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
				repo.On("UpdatePasswordById", mock.Anything, mock.MatchedBy(func(acc auth.Account) bool {
					return acc.UserId == 1 && len(acc.Password) > 0
				})).Return(&auth.Account{UserId: 1}, nil)
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
			name:      "error - invalid JSON",
			body:      `{"password":"test"`,
			userId:    1,
			setUserId: true,
			mockRepo:  func(repo *MockAuthRepository) {},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr: true,
		},
		{
			name:      "error - missing userId in context",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			setUserId: false,
			mockRepo:  func(repo *MockAuthRepository) {},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "Invalid user id",
		},
		{
			name:      "error - current password validation fails",
			body:      `{"password":"wrongpassword","newpassword":"newpassword123"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "incorrect password",
		},
		{
			name:      "error - user not found",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:    999,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 999).
					Return("", errors.New("no user found"))
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "invalid account",
		},
		{
			name:      "error - transaction begin fails",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
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
			name:      "error - new password too short",
			body:      `{"password":"oldpassword123","newpassword":"1234"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectRollback()
				return db, mock
			},
			expectErr:     true,
			errorContains: "invalid password",
		},
		{
			name:      "error - change password service fails",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
				repo.On("UpdatePasswordById", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error"))
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectRollback()
				return db, mock
			},
			expectErr:     true,
			errorContains: "database error",
		},
		{
			name:      "error - transaction commit fails",
			body:      `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
				repo.On("UpdatePasswordById", mock.Anything, mock.Anything).
					Return(&auth.Account{UserId: 1}, nil)
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
			name:      "success - minimum valid new password length",
			body:      `{"password":"oldpassword123","newpassword":"12345"}`,
			userId:    1,
			setUserId: true,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPasswordById", mock.Anything, 1).Return(validOldPasswordHash, nil)
				repo.On("UpdatePasswordById", mock.Anything, mock.Anything).
					Return(&auth.Account{UserId: 1}, nil)
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
			mockRepo := &MockAuthRepository{}
			tt.mockRepo(mockRepo)

			authService := auth.NewAuthService(mockRepo)

			db, sqlMock := tt.mockDB()
			defer db.Close()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/change-password", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.setUserId {
				c.Set("userId", tt.userId)
			}

			err := utils.ChangePasswordById(c, authService, db)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if err := sqlMock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled sqlmock expectations: %s", err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test Login Function.
func TestLoginJwt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Bcrypt hash of "oldpassword123" (reusing from TestChangePasswordById)
	validPasswordHash := "$2a$10$tNMPSwPNfwwKmL0i7J00beidHBjxX9gmyESatFe5GBlxETjwRhMNm"

	tests := []struct {
		name          string
		body          string
		mockRepo      func(repo *MockAuthRepository)
		expectErr     bool
		errorContains string
		validateToken bool
	}{
		{
			name: "success - valid login",
			body: `{"username":"testuser","password":"oldpassword123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "testuser").
					Return(1, validPasswordHash, nil)
			},
			expectErr:     false,
			validateToken: true,
		},
		{
			name: "error - invalid JSON",
			body: `{"username":"test"`,
			mockRepo: func(repo *MockAuthRepository) {
				// No mock needed as it should fail before repository call
			},
			expectErr: true,
		},
		{
			name: "error - missing username",
			body: `{"password":"oldpassword123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "").
					Return(0, "", errors.New("user not found"))
			},
			expectErr: true,
		},
		{
			name: "error - missing password",
			body: `{"username":"testuser"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "testuser").
					Return(1, validPasswordHash, nil)
			},
			expectErr: true,
		},
		{
			name: "error - empty request body",
			body: `{}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "").
					Return(0, "", errors.New("user not found"))
			},
			expectErr: true,
		},
		{
			name: "error - user not found",
			body: `{"username":"nonexistent","password":"oldpassword123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "nonexistent").
					Return(0, "", errors.New("user not found"))
			},
			expectErr:     true,
			errorContains: "cannot retrieve account",
		},
		{
			name: "error - incorrect password",
			body: `{"username":"testuser","password":"wrongpassword"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "testuser").
					Return(1, validPasswordHash, nil)
			},
			expectErr:     true,
			errorContains: "incorrect password",
		},
		{
			name: "success - valid login with different user",
			body: `{"username":"anotheruser","password":"oldpassword123"}`,
			mockRepo: func(repo *MockAuthRepository) {
				repo.On("GetUsrPassword", mock.Anything, "anotheruser").
					Return(2, validPasswordHash, nil)
			},
			expectErr:     false,
			validateToken: true,
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
			db, _, _ := sqlmock.New()
			defer db.Close()

			// Create Gin context with test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute the function
			tokenReturn, err := utils.LoginJwt(c, authService, db)

			// Assertions
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Validate token response for success cases
				if tt.validateToken {
					assert.NotZero(t, tokenReturn.UserId, "UserId should not be zero")
					assert.NotEmpty(t, tokenReturn.AccessToken, "AccessToken should not be empty")
					assert.NotEmpty(t, tokenReturn.TokenId, "TokenId should not be empty")

					// Verify cookies were set
					cookies := w.Result().Cookies()
					var accessTokenFound, refreshTokenFound bool
					for _, cookie := range cookies {
						if cookie.Name == "access_token" {
							accessTokenFound = true
							assert.NotEmpty(t, cookie.Value)
							assert.Equal(t, "/", cookie.Path)
							assert.True(t, cookie.HttpOnly)
							assert.Equal(t, 600, cookie.MaxAge)
						}
						if cookie.Name == "refresh_token" {
							refreshTokenFound = true
							assert.NotEmpty(t, cookie.Value)
							assert.Equal(t, "/", cookie.Path)
							assert.True(t, cookie.HttpOnly)
							assert.Equal(t, 604800, cookie.MaxAge)
						}
					}
					assert.True(t, accessTokenFound, "access_token cookie should be set")
					assert.True(t, refreshTokenFound, "refresh_token cookie should be set")
				}
			}

			// Verify all repository mock expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

// Test Refresh
func TestRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create valid refresh tokens for testing
	validRefreshToken, _ := authen.CreateRefreshToken(1)
	validRefreshToken2, _ := authen.CreateRefreshToken(2)

	tests := []struct {
		name          string
		setupCookie   func(c *gin.Context)
		expectErr     bool
		errorContains string
		validateToken bool
	}{
		{
			name: "success - valid refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: validRefreshToken,
				})
			},
			expectErr:     false,
			validateToken: true,
		},
		{
			name: "error - missing refresh token",
			setupCookie: func(c *gin.Context) {
				// Don't set any cookie
			},
			expectErr:     true,
			errorContains: "missing refresh token",
		},
		{
			name: "error - invalid refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "invalid.token.here",
				})
			},
			expectErr:     true,
			errorContains: "invalid refresh token",
		},
		{
			name: "error - expired refresh token",
			setupCookie: func(c *gin.Context) {
				// Create an expired token manually
				expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
					jwt.MapClaims{
						"userId": 1,
						"exp":    time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
						"type":   "refresh",
					})
				expiredTokenString, _ := expiredToken.SignedString([]byte(serverutils.GetConfigJWT()))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: expiredTokenString,
				})
			},
			expectErr:     true,
			errorContains: "invalid refresh token",
		},
		{
			name: "error - malformed refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "malformed-token-without-proper-structure",
				})
			},
			expectErr:     true,
			errorContains: "invalid refresh token",
		},
		{
			name: "success - valid refresh token for different user",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: validRefreshToken2,
				})
			},
			expectErr:     false,
			validateToken: true,
		},
		{
			name: "error - refresh token with invalid userId type",
			setupCookie: func(c *gin.Context) {
				// Create a token with userId as string instead of int
				invalidToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
					jwt.MapClaims{
						"userId": "invalid-string-id",
						"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
						"type":   "refresh",
					})
				invalidTokenString, _ := invalidToken.SignedString([]byte(serverutils.GetConfigJWT()))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: invalidTokenString,
				})
			},
			expectErr:     true,
			errorContains: "invalid userID type",
		},
		{
			name: "error - refresh token without userId claim",
			setupCookie: func(c *gin.Context) {
				// Create a token without userId claim
				tokenWithoutUserId := jwt.NewWithClaims(jwt.SigningMethodHS256,
					jwt.MapClaims{
						"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
						"type": "refresh",
					})
				tokenString, _ := tokenWithoutUserId.SignedString([]byte(serverutils.GetConfigJWT()))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tokenString,
				})
			},
			expectErr:     true,
			errorContains: "invalid userID type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Gin context with test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/refresh", nil)

			// Setup cookie for this test case
			tt.setupCookie(c)

			// Execute the function
			err := utils.Refresh(c)

			// Assertions
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Validate that new access token was set for success cases
				if tt.validateToken {
					cookies := w.Result().Cookies()
					var accessTokenFound bool
					for _, cookie := range cookies {
						if cookie.Name == "access_token" {
							accessTokenFound = true
							assert.NotEmpty(t, cookie.Value)
							assert.Equal(t, "/", cookie.Path)
							assert.True(t, cookie.HttpOnly)
							assert.Equal(t, 600, cookie.MaxAge) // 10 minutes
						}
					}
					assert.True(t, accessTokenFound, "access_token cookie should be set after successful refresh")
				}
			}
		})
	}
}
