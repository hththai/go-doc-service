package authApi_test

import (
	"2_Go/api/v1/authApi"
	"2_Go/internal/auth"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	auth.AuthService
	mock.Mock
}

func (m *MockAuthService) ChangePassword(c *gin.Context, db *sql.DB) error {
	args := m.Called(c, db)
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
	mockSvc := new(MockAuthService)

	h := &authApi.AuthHandler{
		AccountSvc: mockSvc,
		Logger:     logger,
		DB:         nil, // Use mock DB if your service requires it
	}

	// 2. Define Mock Expectation
	mockSvc.On("ChangePassword", mock.Anything, mock.Anything).Return(nil)

	// 3. Create Request and Inject URL Params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set the URL parameter ":username" manually for the test
	c.Params = []gin.Param{{Key: "username", Value: "johndoe"}}

	// Create a dummy JSON body (if your v1.ChangePassword binds a body)
	reqBody := map[string]string{"new_password": "securepassword123"}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/users/johndoe/changepassword", bytes.NewReader(bodyBytes))
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

// Mock the validation method
func (m *MockAuthService) IsValidToken(c *gin.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

// func TestGetMeHandler(t *testing.T) {
// 	// 1. Setup
// 	gin.SetMode(gin.TestMode)
// 	logger, _ := test.NewNullLogger()
// 	mockSvc := new(MockAuthService)

// 	h := &authApi.AuthHandler{
// 		AccountSvc: mockSvc,
// 		Logger:     logger,
// 	}

// 	// 2. Define expectation for IsValidToken
// 	expectedUser := "test-user-001"
// 	mockSvc.On("IsValidToken", mock.Anything).Return(expectedUser, nil)

// 	// 3. Create the request with the JSON body
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)

// 	// The handler expects this JSON struct
// 	reqBody := map[string]string{"tokenId": expectedUser}
// 	bodyBytes, _ := json.Marshal(reqBody)

// 	c.Request, _ = http.NewRequest("GET", "/users/me", bytes.NewReader(bodyBytes))
// 	c.Request.Header.Set("Content-Type", "application/json")

// 	// 4. Run the handler
// 	h.GetMe(c)

// 	// 5. Assertions
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), expectedUser)
// 	mockSvc.AssertExpectations(t)
// }

func TestHandleRefreshSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.RemoteAddr = "127.0.0.1:8088"
	c.Request = req

	logger, _ := test.NewNullLogger()
	mockSvc := new(MockAuthService)

	handler := &authApi.AuthHandler{
		Logger:     logger, // stub logger
		AccountSvc: mockSvc,
	}

	handler.HandleRefresh(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "Success", body["message"])
}
