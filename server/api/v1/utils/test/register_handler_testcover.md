# Register Handler Test Coverage

This document describes the comprehensive unit test suite for the `Register` function in the authentication handler.

## Overview

The test suite provides complete coverage of the `Register` function, testing all success paths and error scenarios. All tests use mocking to isolate the function behavior and avoid external dependencies.

## Test Statistics

- **Total Test Cases**: 11
- **Success Scenarios**: 3
- **Error Scenarios**: 8
- **Test Status**: ✅ All Passing

## Success Test Cases

### 1. Valid Registration
**Test Name**: `success - valid registration`

Tests the happy path with valid username and password.

**Input**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Expected**: No error, transaction committed successfully

---

### 2. Valid Registration with Email
**Test Name**: `success - valid registration with email`

Tests registration with optional email field included.

**Input**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**Expected**: No error, all fields properly processed

---

### 3. Minimum Valid Password Length
**Test Name**: `success - minimum valid password length`

Tests edge case with exactly 5 characters (minimum required length).

**Input**:
```json
{
  "username": "testuser",
  "password": "12345"
}
```

**Expected**: No error, password meets minimum requirement

---

## Error Test Cases

### 4. Invalid JSON
**Test Name**: `error - invalid JSON`

Tests handling of malformed JSON in request body.

**Input**:
```json
{"username":"test"
```

**Expected**: Error before validation, transaction not started

---

### 5. Missing Username
**Test Name**: `error - missing username`

Tests validation when required username field is missing.

**Input**:
```json
{
  "password": "password123"
}
```

**Expected**: Validation error containing "Username"

---

### 6. Password Too Short
**Test Name**: `error - password too short`

Tests validation when password is less than 5 characters.

**Input**:
```json
{
  "username": "testuser",
  "password": "1234"
}
```

**Expected**: Validation error containing "Password"

---

### 7. Missing Password
**Test Name**: `error - missing password`

Tests validation when required password field is missing.

**Input**:
```json
{
  "username": "testuser"
}
```

**Expected**: Validation error containing "Password"

---

### 8. Empty Request Body
**Test Name**: `error - empty request body`

Tests validation when both required fields are missing.

**Input**:
```json
{}
```

**Expected**: Validation error for missing fields

---

### 9. Transaction Begin Fails
**Test Name**: `error - transaction begin fails`

Tests error handling when database transaction initialization fails.

**Input**: Valid registration data

**Mock Behavior**: Database returns error on `Begin()`

**Expected**: Error containing "failed to start transaction"

---

### 10. Repository Register Fails
**Test Name**: `error - repository register fails (username exists)`

Tests error handling when username already exists in database.

**Input**:
```json
{
  "username": "existinguser",
  "password": "password123"
}
```

**Mock Behavior**: Repository returns "username exists" error

**Expected**: Error containing "username exists", transaction rolled back

---

### 11. Transaction Commit Fails
**Test Name**: `error - transaction commit fails`

Tests error handling when database commit operation fails.

**Input**: Valid registration data

**Mock Behavior**: Database returns error on `Commit()`

**Expected**: Error containing "commit failed"

---

## Implementation Details

### Testing Stack

- **Test Framework**: Go's built-in `testing` package
- **Assertion Library**: `github.com/stretchr/testify/assert`
- **Mock Library**: `github.com/stretchr/testify/mock`
- **SQL Mock**: `github.com/DATA-DOG/go-sqlmock`
- **HTTP Mock**: `github.com/gin-gonic/gin` test mode with `httptest`

### Mocking Strategy

The tests use a layered mocking approach:

1. **Mock AuthRepository**: Custom mock implementing the `auth.AuthRepository` interface
   - All repository methods are mocked
   - Allows control over database operation outcomes

2. **Real AuthService**: Uses `auth.NewAuthService(mockRepo)`
   - Tests the actual service logic including password hashing
   - Avoids complexity of mocking unexported interface methods

3. **Mock Database**: Uses `go-sqlmock` for SQL operations
   - Mocks `Begin()`, `Commit()`, and `Rollback()` operations
   - Verifies proper transaction handling

4. **Mock HTTP Context**: Uses `gin.CreateTestContext()`
   - Simulates HTTP requests with test data
   - Allows testing of JSON binding and validation

### Test Structure

Each test case follows this structure:

```go
{
    name: "test case name",
    body: "JSON request body",
    mockRepo: func(repo *MockAuthRepository) {
        // Setup repository mock expectations
    },
    mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
        // Setup database mock expectations
    },
    expectErr: true/false,
    errorContains: "expected error message substring",
}
```

### Assertions

Each test verifies:

1. **Error State**: Whether an error occurred as expected
2. **Error Message**: Contains expected error text (for error cases)
3. **SQL Expectations**: All database operations were called as expected
4. **Mock Expectations**: All repository methods were called correctly

## Running the Tests

```bash
# Run all tests in the package
go test -v ./api/v1/utils/test/

# Run only the Register tests
go test -v ./api/v1/utils/test/ -run TestRegister

# Run with coverage
go test -v -cover ./api/v1/utils/test/ -run TestRegister
```

## Test Results

```
=== RUN   TestRegister
=== RUN   TestRegister/success_-_valid_registration
=== RUN   TestRegister/error_-_invalid_JSON
=== RUN   TestRegister/error_-_missing_username
=== RUN   TestRegister/error_-_password_too_short
=== RUN   TestRegister/error_-_missing_password
=== RUN   TestRegister/error_-_empty_request_body
=== RUN   TestRegister/error_-_transaction_begin_fails
=== RUN   TestRegister/error_-_repository_register_fails_(username_exists)
=== RUN   TestRegister/error_-_transaction_commit_fails
=== RUN   TestRegister/success_-_valid_registration_with_email
=== RUN   TestRegister/success_-_minimum_valid_password_length
--- PASS: TestRegister (0.25s)
    --- PASS: TestRegister/success_-_valid_registration (0.07s)
    --- PASS: TestRegister/error_-_invalid_JSON (0.00s)
    --- PASS: TestRegister/error_-_missing_username (0.00s)
    --- PASS: TestRegister/error_-_password_too_short (0.00s)
    --- PASS: TestRegister/error_-_missing_password (0.00s)
    --- PASS: TestRegister/error_-_empty_request_body (0.00s)
    --- PASS: TestRegister/error_-_transaction_begin_fails (0.00s)
    --- PASS: TestRegister/error_-_repository_register_fails_(username_exists) (0.05s)
    --- PASS: TestRegister/error_-_transaction_commit_fails (0.05s)
    --- PASS: TestRegister/success_-_valid_registration_with_email (0.05s)
    --- PASS: TestRegister/success_-_minimum_valid_password_length (0.04s)
PASS
```

## Coverage Summary

The test suite covers:

- ✅ JSON binding and validation
- ✅ Account field validation (username, password)
- ✅ Password length requirements (minimum 5 characters)
- ✅ Database transaction management (begin, commit, rollback)
- ✅ Service layer error handling
- ✅ Repository layer error handling
- ✅ Edge cases (minimum valid inputs)
- ✅ Error propagation through all layers

## Future Enhancements

Potential additions to the test suite:

1. Test concurrent registration attempts
2. Test with various special characters in username/password
3. Test with extremely long username/password inputs
4. Test GUID generation uniqueness
5. Performance/benchmark tests
6. Integration tests with real database

---

**Last Updated**: 2026-01-31
**Test File**: [register_handler_test.go](./register_handler_test.go)
