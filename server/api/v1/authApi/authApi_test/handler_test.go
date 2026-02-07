package authApi_test

import (
	"2_Go/api/v1/authApi"
	"2_Go/internal/auth"
	"2_Go/middleware/authen"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	auth.AuthService
	mock.Mock
}

func (m *MockAuthService) ValidateAccountByIdService(db *sql.DB, userId int, password string) error {
	args := m.Called(db, userId, password)
	return args.Error(0)
}

func (m *MockAuthService) ChangePasswordService(tx *sql.Tx, account auth.Account, newPassword string) (*auth.Account, error) {
	args := m.Called(tx, account, newPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Account), args.Error(1)
}

// Mock the validation method
func (m *MockAuthService) IsValidToken(c *gin.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

// New interface methods

func (m *MockAuthService) ChangePassword(db *sql.DB, userId int, currentPassword, newPassword string) error {
	args := m.Called(db, userId, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) RefreshAccessToken(refreshToken string) (newAccessToken, newTokenId string, userId int, err error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Int(2), args.Error(3)
}

func (m *MockAuthService) ValidateAccessToken(accessToken string) (tokenId string, err error) {
	args := m.Called(accessToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CreateTokensForUser(userId int) (accessToken, tokenId, refreshToken string, err error) {
	args := m.Called(userId)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) RegisterAccount(db *sql.DB, account auth.Account) error {
	args := m.Called(db, account)
	return args.Error(0)
}

// Test GetPing
func TestGetPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1. Create a null logger that satisfies the FieldLogger interface
	nullLogger, _ := test.NewNullLogger()

	h := &authApi.AuthHandler{
		Logger: nullLogger, // Works perfectly with FieldLogger
	}

	r.GET("/ping", h.GetPing)

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"success"}`, w.Body.String())
}

func TestHandleChangePasswordSuccess(t *testing.T) {
	// 1. Setup
	gin.SetMode(gin.TestMode)
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel // Enable debug level logging
	mockSvc := new(MockAuthService)

	// Create a sqlmock database
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	h := &authApi.AuthHandler{
		AccountSvc: mockSvc,
		Logger:     logger,
		DB:         db,
	}

	// 2. Define Mock Expectations - now uses ChangePassword
	mockSvc.On("ChangePassword", db, 123, "oldpassword123", "securepassword123").Return(nil)

	// 3. Create Request and Inject URL Params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set the URL parameter ":username" manually for the test
	c.Params = []gin.Param{{Key: "username", Value: "johndoe"}}

	// Set context values required by the handler
	c.Set("username", "johndoe") // Required by utils.IsValidUsername
	c.Set("userId", 123)         // Required by handler

	// Create a JSON body with password and newpassword
	reqBody := map[string]string{
		"password":    "oldpassword123",
		"newpassword": "securepassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/user/changepassword", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// 4. Execute
	h.HandleChangePassword(c)

	// 5. Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify structured logging via FieldLogger
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, "debug", hook.LastEntry().Level.String())
	assert.Contains(t, hook.LastEntry().Message, "password updated success")

	mockSvc.AssertExpectations(t)
}

func TestHandleRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a valid refresh token for testing
	testUserId := 123
	refreshToken, err := authen.CreateRefreshToken(testUserId)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.RemoteAddr = "127.0.0.1:8088"

	// Set the refresh_token cookie in the request
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})

	c.Request = req

	logger, _ := test.NewNullLogger()
	mockSvc := new(MockAuthService)

	// Mock the RefreshAccessToken method
	mockSvc.On("RefreshAccessToken", refreshToken).Return("new_access_token", "new_token_id", testUserId, nil)

	handler := &authApi.AuthHandler{
		Logger:     logger,
		AccountSvc: mockSvc,
	}

	handler.HandleRefresh(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "Success", body["message"])

	// Verify that a new access_token cookie was set
	cookies := w.Result().Cookies()
	var accessTokenFound bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			accessTokenFound = true
			assert.NotEmpty(t, cookie.Value)
			break
		}
	}
	assert.True(t, accessTokenFound, "access_token cookie should be set")

	mockSvc.AssertExpectations(t)
}
