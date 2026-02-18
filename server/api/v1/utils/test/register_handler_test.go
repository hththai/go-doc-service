package test

import (
	"2_Go/api/v1/authApi"
	"2_Go/internal/auth"
	"2_Go/internal/config"
	"2_Go/middleware/authen"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	config.Load()
}

// MockAuthService mocks the auth.AuthService interface
// Embedding auth.AuthService allows the mock to satisfy the interface
// including unexported methods
type MockAuthService struct {
	auth.AuthService
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (*auth.Account, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

func (m *MockAuthService) CreateTokensForUser(userId int) (accessToken, tokenId, refreshToken string, err error) {
	args := m.Called(userId)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) RefreshAccessToken(refreshToken string) (newAccessToken, newTokenId string, userId int, err error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Int(2), args.Error(3)
}

func (m *MockAuthService) RegisterAccount(account auth.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(userId int, currentPassword, newPassword string) error {
	args := m.Called(userId, currentPassword, newPassword)
	return args.Error(0)
}

// TestHandleRegister tests the HandleRegister handler method
func TestHandleRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *MockAuthService)
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name: "success - valid registration",
			body: `{"username":"testuser","password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.MatchedBy(func(acc auth.Account) bool {
					return acc.Username == "testuser" && acc.Password == "password123"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "error - invalid JSON",
			body: `{"username":"test"`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "error - missing username",
			body: `{"password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed - Gin binding validation fails before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Username",
		},
		{
			name: "error - password too short",
			body: `{"username":"testuser","password":"1234"}`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed - Gin binding validation fails before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Password",
		},
		{
			name: "error - missing password",
			body: `{"username":"testuser"}`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed - Gin binding validation fails before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Password",
		},
		{
			name: "error - empty request body",
			body: `{}`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed - Gin binding validation fails before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Username",
		},
		{
			name: "error - transaction begin fails",
			body: `{"username":"testuser","password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.Anything).
					Return(errors.New("failed to start transaction"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "failed to start transaction",
		},
		{
			name: "error - username exists",
			body: `{"username":"existinguser","password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.Anything).
					Return(errors.New("username exists"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "username exists",
		},
		{
			name: "error - transaction commit fails",
			body: `{"username":"testuser","password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.Anything).
					Return(errors.New("commit failed"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "commit failed",
		},
		{
			name: "success - valid registration with email",
			body: `{"username":"testuser","email":"test@example.com","password":"password123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.MatchedBy(func(acc auth.Account) bool {
					return acc.Username == "testuser" && acc.Email == "test@example.com"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "success - minimum valid password length",
			body: `{"username":"testuser","password":"12345"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("RegisterAccount", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockSvc := new(MockAuthService)

			// Setup mock expectations
			tt.mockSetup(mockSvc)

			// Create logger
			logger, _ := test.NewNullLogger()

			// Create handler
			handler := &authApi.AuthHandler{
				AccountSvc: mockSvc,
				Logger:     logger,
			}

			// Create Gin context with test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute the handler
			handler.HandleRegister(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError && tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}

			// Verify all mock expectations were met
			mockSvc.AssertExpectations(t)
		})
	}
}

// TestHandleChangePassword tests the HandleChangePassword handler method
func TestHandleChangePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		userId         interface{}
		setUserId      bool
		setUsername    bool
		mockSetup      func(svc *MockAuthService)
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:        "success - valid password change",
			body:        `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "newpassword123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "error - invalid JSON",
			body:           `{"password":"test"`,
			userId:         1,
			setUserId:      true,
			setUsername:    true,
			mockSetup:      func(svc *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "error - missing userId in context",
			body:           `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			setUserId:      false,
			setUsername:    true,
			mockSetup:      func(svc *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "Invalid user id",
		},
		{
			name:        "error - current password validation fails",
			body:        `{"password":"wrongpassword","newpassword":"newpassword123"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "wrongpassword", "newpassword123").
					Return(errors.New("incorrect password"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "incorrect password",
		},
		{
			name:        "error - user not found",
			body:        `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:      999,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 999, "oldpassword123", "newpassword123").
					Return(errors.New("invalid account"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid account",
		},
		{
			name:        "error - transaction begin fails",
			body:        `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "newpassword123").
					Return(errors.New("failed to start transaction"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "failed to start transaction",
		},
		{
			name:        "error - new password too short",
			body:        `{"password":"oldpassword123","newpassword":"1234"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "1234").
					Return(errors.New("invalid password criteria"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid password",
		},
		{
			name:        "error - change password service fails",
			body:        `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "newpassword123").
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "database error",
		},
		{
			name:        "error - transaction commit fails",
			body:        `{"password":"oldpassword123","newpassword":"newpassword123"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "newpassword123").
					Return(errors.New("commit failed"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "commit failed",
		},
		{
			name:        "success - minimum valid new password length",
			body:        `{"password":"oldpassword123","newpassword":"12345"}`,
			userId:      1,
			setUserId:   true,
			setUsername: true,
			mockSetup: func(svc *MockAuthService) {
				svc.On("ChangePassword", 1, "oldpassword123", "12345").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)

			tt.mockSetup(mockSvc)

			logger, _ := test.NewNullLogger()

			handler := &authApi.AuthHandler{
				AccountSvc: mockSvc,
				Logger:     logger,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/user/changepassword", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Set URL params for username validation
			c.Params = []gin.Param{{Key: "username", Value: "testuser"}}

			if tt.setUsername {
				c.Set("username", "testuser")
			}

			if tt.setUserId {
				c.Set("userId", tt.userId)
			}

			handler.HandleChangePassword(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError && tt.errorContains != "" {
				assert.Contains(t, w.Body.String(), tt.errorContains)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

// TestHandleLogin tests the HandleLogin handler method
func TestHandleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *MockAuthService)
		expectedStatus int
		expectError    bool
		errorContains  string
		validateToken  bool
	}{
		{
			name: "success - valid login",
			body: `{"username":"testuser","password":"oldpassword123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "testuser", "oldpassword123").
					Return(&auth.Account{UserId: 1, Username: "testuser"}, nil)
				svc.On("CreateTokensForUser", 1).
					Return("access_token_value", "token_id_value", "refresh_token_value", nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateToken:  true,
		},
		{
			name: "error - invalid JSON",
			body: `{"username":"test"`,
			mockSetup: func(svc *MockAuthService) {
				// No mock needed as it should fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "error - missing username",
			body: `{"password":"oldpassword123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "", "oldpassword123").
					Return(nil, errors.New("cannot retrieve account"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "error - missing password",
			body: `{"username":"testuser"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "testuser", "").
					Return(nil, errors.New("incorrect password"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "error - empty request body",
			body: `{}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "", "").
					Return(nil, errors.New("cannot retrieve account"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "error - user not found",
			body: `{"username":"nonexistent","password":"oldpassword123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "nonexistent", "oldpassword123").
					Return(nil, errors.New("cannot retrieve account"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "cannot retrieve account",
		},
		{
			name: "error - incorrect password",
			body: `{"username":"testuser","password":"wrongpassword"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "testuser", "wrongpassword").
					Return(nil, errors.New("incorrect password"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "incorrect password",
		},
		{
			name: "error - token creation fails",
			body: `{"username":"testuser","password":"oldpassword123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "testuser", "oldpassword123").
					Return(&auth.Account{UserId: 1, Username: "testuser"}, nil)
				svc.On("CreateTokensForUser", 1).
					Return("", "", "", errors.New("failed to create session"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "failed to create session",
		},
		{
			name: "success - valid login with different user",
			body: `{"username":"anotheruser","password":"oldpassword123"}`,
			mockSetup: func(svc *MockAuthService) {
				svc.On("Login", "anotheruser", "oldpassword123").
					Return(&auth.Account{UserId: 2, Username: "anotheruser"}, nil)
				svc.On("CreateTokensForUser", 2).
					Return("access_token_value_2", "token_id_value_2", "refresh_token_value_2", nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateToken:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)

			tt.mockSetup(mockSvc)

			logger, _ := test.NewNullLogger()

			handler := &authApi.AuthHandler{
				AccountSvc: mockSvc,
				Logger:     logger,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.HandleLogin(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				if tt.errorContains != "" {
					assert.Contains(t, w.Body.String(), tt.errorContains)
				}
			} else if tt.validateToken {
				// Verify response contains token info
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["access_token"])
				assert.NotEmpty(t, response["tokenId"])

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

			mockSvc.AssertExpectations(t)
		})
	}
}

// TestHandleRefresh tests the HandleRefresh handler method
func TestHandleRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create valid refresh tokens for testing
	validRefreshToken, _ := authen.CreateRefreshToken(1)
	validRefreshToken2, _ := authen.CreateRefreshToken(2)

	tests := []struct {
		name           string
		setupCookie    func(c *gin.Context)
		mockSetup      func(svc *MockAuthService)
		expectedStatus int
		expectError    bool
		errorContains  string
		validateToken  bool
	}{
		{
			name: "success - valid refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: validRefreshToken,
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", validRefreshToken).
					Return("new_access_token", "new_token_id", 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateToken:  true,
		},
		{
			name: "error - missing refresh token",
			setupCookie: func(c *gin.Context) {
				// Don't set any cookie
			},
			mockSetup:      func(svc *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "missing refresh token",
		},
		{
			name: "error - invalid refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "invalid.token.here",
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", "invalid.token.here").
					Return("", "", 0, errors.New("invalid refresh token"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid refresh token",
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
				expiredTokenString, _ := expiredToken.SignedString([]byte(config.Get().JWT.Secret))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: expiredTokenString,
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", mock.Anything).
					Return("", "", 0, errors.New("invalid refresh token"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid refresh token",
		},
		{
			name: "error - malformed refresh token",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "malformed-token-without-proper-structure",
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", "malformed-token-without-proper-structure").
					Return("", "", 0, errors.New("invalid refresh token"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid refresh token",
		},
		{
			name: "success - valid refresh token for different user",
			setupCookie: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: validRefreshToken2,
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", validRefreshToken2).
					Return("new_access_token_2", "new_token_id_2", 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateToken:  true,
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
				invalidTokenString, _ := invalidToken.SignedString([]byte(config.Get().JWT.Secret))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: invalidTokenString,
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", mock.Anything).
					Return("", "", 0, errors.New("invalid userId type in token"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid userId type",
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
				tokenString, _ := tokenWithoutUserId.SignedString([]byte(config.Get().JWT.Secret))
				c.Request.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tokenString,
				})
			},
			mockSetup: func(svc *MockAuthService) {
				svc.On("RefreshAccessToken", mock.Anything).
					Return("", "", 0, errors.New("invalid userId type in token"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid userId type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)
			tt.mockSetup(mockSvc)

			logger, _ := test.NewNullLogger()

			handler := &authApi.AuthHandler{
				AccountSvc: mockSvc,
				Logger:     logger,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/refresh", nil)

			// Setup cookie for this test case
			tt.setupCookie(c)

			handler.HandleRefresh(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				if tt.errorContains != "" {
					assert.Contains(t, w.Body.String(), tt.errorContains)
				}
			} else if tt.validateToken {
				// Validate that new access token was set for success cases
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

			mockSvc.AssertExpectations(t)
		})
	}
}
