# Authentication Handler Test Coverage

This document describes the comprehensive unit test suites for authentication handler functions including `Register` and `ChangePasswordById`.

## Overview

The test suites provide complete coverage of authentication functions, testing all success paths and error scenarios. All tests use mocking to isolate function behavior and avoid external dependencies.

## Overall Test Statistics

- **Total Test Cases**: 21 (11 Register + 10 ChangePasswordById)
- **Success Scenarios**: 5
- **Error Scenarios**: 16
- **Test Status**: ✅ All Passing

---

# Register Function Tests

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

# ChangePasswordById Function Tests

## Test Statistics

- **Total Test Cases**: 10
- **Success Scenarios**: 2
- **Error Scenarios**: 8
- **Test Status**: ✅ All Passing

## Success Test Cases

### 1. Valid Password Change
**Test Name**: `success - valid password change`

Tests the happy path where a user successfully changes their password with correct current password.

**Input**:
```json
{
  "password": "oldpassword123",
  "newpassword": "newpassword123"
}
```

**Context**: userId = 1 (set by authentication middleware)

**Expected**: No error, password updated successfully, transaction committed

---

### 2. Minimum Valid New Password Length
**Test Name**: `success - minimum valid new password length`

Tests edge case with exactly 5 characters for the new password (minimum required length).

**Input**:
```json
{
  "password": "oldpassword123",
  "newpassword": "12345"
}
```

**Context**: userId = 1

**Expected**: No error, new password meets minimum requirement

---

## Error Test Cases

### 3. Invalid JSON
**Test Name**: `error - invalid JSON`

Tests handling of malformed JSON in request body.

**Input**:
```json
{"password":"test"
```

**Expected**: Error before validation, no database operations

---

### 4. Missing userId in Context
**Test Name**: `error - missing userId in context`

Tests error handling when authentication middleware fails to set userId in context.

**Input**: Valid password change request

**Context**: No userId set

**Expected**: Error containing "Invalid user id"

---

### 5. Current Password Validation Fails
**Test Name**: `error - current password validation fails`

Tests validation when the provided current password is incorrect.

**Input**:
```json
{
  "password": "wrongpassword",
  "newpassword": "newpassword123"
}
```

**Context**: userId = 1

**Mock Behavior**: Repository returns hashed password that doesn't match

**Expected**: Error containing "incorrect password"

---

### 6. User Not Found
**Test Name**: `error - user not found`

Tests error handling when the userId doesn't exist in database.

**Input**:
```json
{
  "password": "oldpassword123",
  "newpassword": "newpassword123"
}
```

**Context**: userId = 999 (non-existent)

**Mock Behavior**: Repository returns "no user found" error

**Expected**: Error containing "invalid account"

---

### 7. Transaction Begin Fails
**Test Name**: `error - transaction begin fails`

Tests error handling when database transaction initialization fails.

**Input**: Valid password change request

**Context**: userId = 1

**Mock Behavior**: Database returns error on `Begin()`

**Expected**: Error containing "failed to start transaction"

---

### 8. New Password Too Short
**Test Name**: `error - new password too short`

Tests validation when new password is less than 5 characters.

**Input**:
```json
{
  "password": "oldpassword123",
  "newpassword": "1234"
}
```

**Context**: userId = 1

**Expected**: Error containing "invalid password", transaction rolled back

---

### 9. Change Password Service Fails
**Test Name**: `error - change password service fails`

Tests error handling when the repository fails to update the password.

**Input**: Valid password change request

**Context**: userId = 1

**Mock Behavior**: Repository returns "database error"

**Expected**: Error containing "database error", transaction rolled back

---

### 10. Transaction Commit Fails
**Test Name**: `error - transaction commit fails`

Tests error handling when database commit operation fails.

**Input**: Valid password change request

**Context**: userId = 1

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

**Register Test Structure:**

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

**ChangePasswordById Test Structure:**

```go
{
    name: "test case name",
    body: "JSON request body",
    userId: interface{},  // The userId to set in context
    setUserId: bool,      // Whether to set userId in context
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

**Special Features for ChangePasswordById:**
- Uses actual bcrypt hash (`$2a$10$tNMPSwPNfwwKmL0i7J00beidHBjxX9gmyESatFe5GBlxETjwRhMNm`) for testing password validation
- Tests Gin context userId presence/absence scenarios
- Validates current password before allowing change

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

# Run only the ChangePasswordById tests
go test -v ./api/v1/utils/test/ -run TestChangePasswordById

# Run both authentication handler tests
go test -v ./api/v1/utils/test/ -run "TestRegister|TestChangePasswordById"

# Run with coverage
go test -v -cover ./api/v1/utils/test/
```

## Test Results

### Register Function Test Results

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

### ChangePasswordById Function Test Results

```
=== RUN   TestChangePasswordById
=== RUN   TestChangePasswordById/success_-_valid_password_change
=== RUN   TestChangePasswordById/error_-_invalid_JSON
=== RUN   TestChangePasswordById/error_-_missing_userId_in_context
=== RUN   TestChangePasswordById/error_-_current_password_validation_fails
=== RUN   TestChangePasswordById/error_-_user_not_found
=== RUN   TestChangePasswordById/error_-_transaction_begin_fails
=== RUN   TestChangePasswordById/error_-_new_password_too_short
=== RUN   TestChangePasswordById/error_-_change_password_service_fails
=== RUN   TestChangePasswordById/error_-_transaction_commit_fails
=== RUN   TestChangePasswordById/success_-_minimum_valid_new_password_length
--- PASS: TestChangePasswordById (0.52s)
    --- PASS: TestChangePasswordById/success_-_valid_password_change (0.11s)
    --- PASS: TestChangePasswordById/error_-_invalid_JSON (0.00s)
    --- PASS: TestChangePasswordById/error_-_missing_userId_in_context (0.00s)
    --- PASS: TestChangePasswordById/error_-_current_password_validation_fails (0.05s)
    --- PASS: TestChangePasswordById/error_-_user_not_found (0.00s)
    --- PASS: TestChangePasswordById/error_-_transaction_begin_fails (0.04s)
    --- PASS: TestChangePasswordById/error_-_new_password_too_short (0.05s)
    --- PASS: TestChangePasswordById/error_-_change_password_service_fails (0.09s)
    --- PASS: TestChangePasswordById/error_-_transaction_commit_fails (0.09s)
    --- PASS: TestChangePasswordById/success_-_minimum_valid_new_password_length (0.09s)
PASS
```

## Coverage Summary

### Register Function Coverage

- ✅ JSON binding and validation
- ✅ Account field validation (username, password)
- ✅ Password length requirements (minimum 5 characters)
- ✅ Database transaction management (begin, commit, rollback)
- ✅ Service layer error handling
- ✅ Repository layer error handling
- ✅ Edge cases (minimum valid inputs)
- ✅ Error propagation through all layers
- ✅ GUID generation for new accounts

### ChangePasswordById Function Coverage

- ✅ JSON binding and validation
- ✅ Gin context userId validation
- ✅ Current password verification with bcrypt
- ✅ New password validation (minimum 5 characters)
- ✅ Authentication middleware integration (userId context)
- ✅ Database transaction management (begin, commit, rollback)
- ✅ Service layer error handling
- ✅ Repository layer error handling
- ✅ User existence validation
- ✅ Password update security (requires current password)
- ✅ Edge cases (minimum valid password length)
- ✅ Error propagation through all layers

## Future Enhancements

### Register Function

1. Test concurrent registration attempts
2. Test with various special characters in username/password
3. Test with extremely long username/password inputs
4. Test GUID generation uniqueness
5. Test duplicate username edge cases
6. Performance/benchmark tests

### ChangePasswordById Function

1. Test password strength requirements
2. Test password history (prevent reuse of old passwords)
3. Test rate limiting for password change attempts
4. Test concurrent password change attempts
5. Test session invalidation after password change
6. Test password change with different authentication methods

### General

1. Integration tests with real database
2. End-to-end tests with actual HTTP requests
3. Security testing (SQL injection, XSS prevention)
4. Load testing for authentication endpoints

---

**Last Updated**: 2026-01-31
**Test File**: [register_handler_test.go](./register_handler_test.go)
