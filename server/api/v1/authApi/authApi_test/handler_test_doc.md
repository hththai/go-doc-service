# AuthAPI Handler Test Documentation

## Overview

This document describes the test suite for the AuthAPI handlers located in `handler_test.go`. The tests verify the behavior of authentication-related HTTP handlers including user authentication, password changes, and token operations.

## Test Environment Setup

### Dependencies
- **Testing Framework**: `testing` (Go standard library)
- **Assertion Library**: `github.com/stretchr/testify/assert`
- **Mocking Library**: `github.com/stretchr/testify/mock`
- **HTTP Testing**: `net/http/httptest`
- **Gin Framework**: `github.com/gin-gonic/gin`
- **Database Mocking**: `github.com/DATA-DOG/go-sqlmock`
- **Logger Testing**: `github.com/sirupsen/logrus/hooks/test`

### Mock Implementations

#### MockAuthService
Mocks the `auth.AuthService` interface with the following methods:

```go
type MockAuthService struct {
    auth.AuthService
    mock.Mock
}
```

**Mocked Methods:**
- `ValidateAccountByIdService(db *sql.DB, userId int, password string) error`
  - Validates user credentials by user ID
  - Returns error if validation fails

- `ChangePasswordService(tx *sql.Tx, account auth.Account, newPassword string) (*auth.Account, error)`
  - Changes the user's password within a transaction
  - Returns updated account or error

- `IsValidToken(c *gin.Context) (string, error)`
  - Validates JWT access tokens from cookies
  - Returns username/tokenId or error

## Test Cases

### 1. TestGetPing

**Purpose**: Verifies the health check endpoint returns a successful response.

**Test Flow:**
1. Create a Gin router with test mode enabled
2. Create an AuthHandler with a null logger
3. Register the `/ping` endpoint
4. Send a GET request to `/ping`
5. Assert:
   - Status code is 200 OK
   - Response body is `{"message":"success"}`

**Key Points:**
- Simple health check with no authentication
- No database or service mocking required
- Verifies basic handler setup and routing

---

### 2. TestHandleChangePasswordSuccess

**Purpose**: Verifies that users can successfully change their password when all validations pass.

**Test Flow:**

#### Setup Phase
1. Enable Gin test mode
2. Create a null logger with **Debug level enabled** (critical for capturing log entries)
3. Create mock auth service
4. Create a mock SQL database using sqlmock
5. Initialize AuthHandler with mocks

#### Mock Expectations
```go
// Service method mocks
mockSvc.On("ValidateAccountByIdService", db, 123, "oldpassword123").Return(nil)
mockSvc.On("ChangePasswordService", mock.Anything, mock.AnythingOfType("auth.Account"), "securepassword123").Return(&auth.Account{}, nil)

// Database transaction mocks
mockDB.ExpectBegin()
mockDB.ExpectCommit()
```

#### Request Setup
1. Create test context with httptest recorder
2. Set URL parameter: `username: "johndoe"`
3. **Set context values** (critical for validation):
   - `c.Set("username", "johndoe")` - Required by `utils.IsValidUsername`
   - `c.Set("userId", 123)` - Required by `ChangePasswordById`
4. Create JSON request body:
   ```json
   {
     "password": "oldpassword123",
     "newpassword": "securepassword123"
   }
   ```

#### Execution & Assertions
1. Call `h.HandleChangePassword(c)`
2. Assert:
   - Status code is 200 OK
   - Logger captured exactly 1 entry
   - Log entry level is "debug"
   - Log message contains "password updated success"
3. Verify all mock expectations were met

**Critical Configuration:**
- **Logger Level**: Must be set to `logrus.DebugLevel` to capture debug log entries
- **Context Values**: Both `username` and `userId` must be set in the Gin context
- **URL Parameter**: The `username` param must match the context username for validation

**Handler Flow:**
1. `utils.IsValidUsername(c)` - Validates that JWT username matches URL parameter
2. `v1.ChangePasswordById(c, h.AccountSvc, h.DB)`:
   - Binds JSON request body
   - Gets `userId` from context (fails if not set)
   - Validates old password via `ValidateAccountByIdService`
   - Begins database transaction
   - Changes password via `ChangePasswordService`
   - Commits transaction
3. Logs success message with client IP
4. Returns 200 OK with `{"message":"Success"}`

**Common Issues & Solutions:**

| Issue | Cause | Solution |
|-------|-------|----------|
| Status 403 Forbidden | Username validation fails | Ensure `c.Set("username", "johndoe")` matches URL param |
| Status 400 "Invalid user id" | userId not in context | Add `c.Set("userId", 123)` before handler call |
| No log entries captured | Logger level too high | Set `logger.Level = logrus.DebugLevel` |
| Nil pointer on `hook.LastEntry()` | No log entries | Fix above issues first, check logger level |
| Mock expectation failure | Wrong method mocked | Use `ValidateAccountByIdService` and `ChangePasswordService`, not `ChangePassword` |

---

### 3. TestHandleRefreshSuccess

**Purpose**: Verifies that the token refresh endpoint returns a success response.

**Test Flow:**
1. Create test context with POST request to `/v1/auth/refresh`
2. Set remote address for client IP tracking
3. Create handler with null logger and mock service
4. Call `HandleRefresh`
5. Assert:
   - Status code is 200 OK
   - Response body contains `{"message":"Success"}`

**Current Status**: ⚠️ This test is currently failing and needs fixes similar to TestHandleChangePasswordSuccess (proper cookie setup, mock expectations, etc.)

---

## Common Testing Patterns

### Setting Up Gin Test Context

```go
gin.SetMode(gin.TestMode)
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)

// Set URL parameters
c.Params = []gin.Param{{Key: "username", Value: "johndoe"}}

// Set context values (from middleware)
c.Set("username", "johndoe")
c.Set("userId", 123)

// Create request with JSON body
reqBody := map[string]string{"key": "value"}
bodyBytes, _ := json.Marshal(reqBody)
c.Request, _ = http.NewRequest("POST", "/endpoint", bytes.NewReader(bodyBytes))
c.Request.Header.Set("Content-Type", "application/json")
```

### Setting Up Logger with Hook

```go
logger, hook := test.NewNullLogger()
logger.Level = logrus.DebugLevel  // Enable debug logging

// After handler execution
assert.Equal(t, 1, len(hook.Entries))
assert.Equal(t, "debug", hook.LastEntry().Level.String())
assert.Contains(t, hook.LastEntry().Message, "expected message")
```

### Setting Up Database Mocks

```go
db, mockDB, err := sqlmock.New()
assert.NoError(t, err)
defer db.Close()

// Expect transaction
mockDB.ExpectBegin()
mockDB.ExpectCommit()

// Verify expectations
assert.NoError(t, mockDB.ExpectationsWereMet())
```

### Setting Up Service Mocks

```go
mockSvc := new(MockAuthService)

// Mock method with specific parameters
mockSvc.On("MethodName", arg1, arg2).Return(returnValue, nil)

// Mock method with any parameters
mockSvc.On("MethodName", mock.Anything, mock.AnythingOfType("TypeName")).Return(value, nil)

// Verify all expectations were met
mockSvc.AssertExpectations(t)
```

## Running Tests

### Run All Tests
```bash
go test -v ./api/v1/authApi/authApi_test
```

### Run Specific Test
```bash
go test -v ./api/v1/authApi/authApi_test -run TestHandleChangePasswordSuccess
```

### Run with Coverage
```bash
go test -v -cover ./api/v1/authApi/authApi_test
```

## Best Practices

1. **Always enable Gin test mode**: `gin.SetMode(gin.TestMode)` to avoid debug output
2. **Set appropriate logger level**: Use `logrus.DebugLevel` when testing log output
3. **Use descriptive test names**: Follow `Test<HandlerName><Scenario>` pattern
4. **Mock all external dependencies**: Database, services, external APIs
5. **Verify all mocks**: Always call `AssertExpectations(t)` and `ExpectationsWereMet()`
6. **Clean up resources**: Use `defer db.Close()` for database connections
7. **Test both success and failure cases**: Include error scenarios
8. **Set realistic context values**: Mimic what middleware would set in production

## File Structure

```
api/v1/authApi/authApi_test/
├── handler_test.go       # Test implementations
└── handler_test_doc.md   # This documentation file
```

## Related Files

- [handler.go](../handler.go) - Handler implementations being tested
- [register_handler.go](../../utils/register_handler.go) - Helper functions used by handlers
- [utils.go](../../../../utils/utils.go) - Validation utilities

## Future Improvements

1. **Add failure test cases**:
   - Invalid credentials
   - Missing required fields
   - Database transaction failures
   - Validation errors

2. **Fix TestHandleRefreshSuccess**:
   - Add proper cookie setup
   - Mock JWT token verification
   - Add service mocks for refresh logic

3. **Add TestGetMeHandler** (currently commented out):
   - Uncomment and fix the test
   - Add proper token validation mocks

4. **Add integration tests**:
   - Test with real database (using testcontainers)
   - Test full authentication flow

## Changelog

### 2026-02-01
- Fixed `TestHandleChangePasswordSuccess`
  - Added correct service method mocks (`ValidateAccountByIdService`, `ChangePasswordService`)
  - Added context value setup (`username`, `userId`)
  - Implemented sqlmock for database transaction testing
  - Set logger to Debug level for log capture
  - Added comprehensive documentation

### Previous
- Initial test implementations for GetPing and HandleRefresh
- Basic mock structure for AuthService
