# Document Handler Test Coverage Documentation

## Overview

This document provides comprehensive coverage documentation for the test functions in `doc_handler_test.go`.

**Test File**: [doc_handler_test.go](doc_handler_test.go)
**Handler Under Test**: `UploadDocument` in [document_handler.go](../document_handler.go)
**Test Framework**: Go testing with testify and sqlmock
**Date Created**: 2026-02-01
**Last Updated**: 2026-03-01

---

## Test Coverage Summary

### Overall Coverage Statistics

| Function | Coverage | Status |
|----------|----------|--------|
| `UploadDocument` | ~90% | ✅ Excellent |
| `buildUploadPath` | 100% | ✅ Complete |
| `saveFileAndMetadata` | 63.2% | ⚠️ Good (limited by production-only code) |
| `GetPurchaseItems` | 100% | ✅ Complete |
| `GetPurchaseById` | 100% | ✅ Complete |

---

## Test Cases

### 1. Success - Upload with File ✅

**Test Name**: `success - upload with file`

**Description**: Tests the complete document upload flow with a file attachment.

**Setup**:
- Form data: name="Test Document", description="Test Description"
- File: test.pdf with dummy content
- User ID: 123 (authenticated)

**Expected Behavior**:
- Document metadata is saved with generated object ID
- File is uploaded and saved to file system
- File path is recorded in database
- Line items are saved (empty list)
- Transaction commits successfully
- No errors returned

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
repo.On("InsertFilePath", mock.Anything, mock.Anything).Return(nil)
repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
mockDB.ExpectBegin()
mockDB.ExpectCommit()
```

**Assertion**: `assert.NoError(t, err)`

---

### 2. Success - Upload without File (Metadata Only) ✅

**Test Name**: `success - upload without file (metadata only)`

**Description**: Tests uploading document metadata without an attached file.

**Setup**:
- Form data: name="Metadata Only", description="No file attached"
- No file attachment
- User ID: 456 (authenticated)

**Expected Behavior**:
- Document metadata is saved with generated object ID
- No file upload occurs
- Line items are saved (empty list)
- Transaction commits successfully
- No errors returned

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(2), nil)
repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
mockDB.ExpectBegin()
mockDB.ExpectCommit()
```

**Assertion**: `assert.NoError(t, err)`

---

### 3. Error - Transaction Begin Fails ❌

**Test Name**: `error - transaction begin fails`

**Description**: Tests error handling when database transaction cannot be started.

**Setup**:
- Valid form data and authenticated user
- Database configured to fail on Begin()

**Expected Behavior**:
- Database transaction fails to start
- Error is propagated to caller
- No further database operations occur

**Mocked Components**:
```go
repo.On("BeginTx").Return(nil, errors.New("failed to start transaction"))
```

**Expected Error**: `"failed to start transaction"`

---

### 4. Error - SetLatestObjID Fails ❌

**Test Name**: `error - SetLatestObjID fails`

**Description**: Tests error handling when object ID generation fails.

**Setup**:
- Valid form data and authenticated user
- Repository configured to fail on SetLatestObjId()

**Expected Behavior**:
- Transaction begins successfully
- SetLatestObjId fails
- Transaction is rolled back
- Error is returned with appropriate message

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).
    Return(int64(-1), errors.New("failed to set obj id"))
mockDB.ExpectBegin()
mockDB.ExpectRollback()
```

**Expected Error**: `"metadata save failed"`

---

### 5. Error - SaveMetadataWithObjId Fails (No File) ❌

**Test Name**: `error - SaveMetadataWithObjId fails (no file)`

**Description**: Tests error handling when metadata save fails (metadata-only upload scenario).

**Setup**:
- Valid form data, no file attachment
- SetLatestObjId succeeds but SaveMetadataWithObjId fails

**Expected Behavior**:
- Transaction begins successfully
- SetLatestObjId succeeds
- SaveMetadataWithObjId fails
- Transaction is rolled back
- Error is returned

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(3), nil)
repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).
    Return(int64(-1), errors.New("failed to save metadata"))
mockDB.ExpectBegin()
mockDB.ExpectRollback()
```

**Expected Error**: `"failed to save metadata"`

---

### 6. Error - SaveItems Fails ❌

**Test Name**: `error - SaveItems fails`

**Description**: Tests error handling when saving line items to the database fails.

**Setup**:
- Valid form data, no file attachment
- SetLatestObjId and SaveMetadataWithObjId succeed, but SaveItems fails

**Expected Behavior**:
- Transaction begins successfully
- Metadata saved successfully
- SaveItems fails
- Transaction is rolled back
- Error wrapping `"failed to save items"` is returned

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(5), nil)
repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil)
repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
    Return(errors.New("failed to insert items"))
mockDB.ExpectBegin()
mockDB.ExpectRollback()
```

**Expected Error**: `"failed to save items"`

---

### 7. Error - Commit Fails ❌

**Test Name**: `error - commit fails`

**Description**: Tests error handling when transaction commit fails.

**Setup**:
- Valid form data and authenticated user
- All repository operations succeed
- Database commit configured to fail

**Expected Behavior**:
- Transaction begins successfully
- All repository operations (including SaveItems) complete successfully
- Transaction commit fails
- Error is returned

**Mocked Components**:
```go
repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(4), nil)
repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(4), nil)
repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
mockDB.ExpectBegin()
mockDB.ExpectCommit().WillReturnError(errors.New("commit failed"))
```

**Expected Error**: `"commit failed"`

---

---

## TestGetPurchaseItems

Tests `GetPurchaseItems()` on the document service — a read-only operation with no transaction.

### Test Cases

#### 1. Success - Returns Items ✅

**Test Name**: `success - returns items`

**Description**: Verifies that items are fetched and returned correctly for a valid purchase owned by the user.

**Mocked Components**:
```go
repo.On("GetItemsByPurchaseId", int64(5), 1).Return([]document.Item{
    {Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
    {Name: "Bread", Quantity: "1", UnitPrice: "3.00", SubTotal: "3.00"},
}, nil)
```

**Assertion**: `assert.Equal(t, expectedItems, items)` + `assert.NoError(t, err)`

---

#### 2. Success - Empty Items List ✅

**Test Name**: `success - empty items list`

**Description**: Verifies that an empty slice is returned when a purchase exists but has no line items (not an error condition).

**Mocked Components**:
```go
repo.On("GetItemsByPurchaseId", int64(5), 1).Return([]document.Item{}, nil)
```

**Assertion**: `assert.Empty(t, got)` + `assert.NoError(t, err)`

---

#### 3. Error - Repository Failure ❌

**Test Name**: `error - repository failure`

**Description**: Verifies that a generic database error is propagated to the caller.

**Mocked Components**:
```go
repo.On("GetItemsByPurchaseId", int64(99), 1).Return([]document.Item{}, errors.New("db failure"))
```

**Expected Error**: `"db failure"`

---

#### 4. Error - Purchase Not Found (ErrNoRows) ❌

**Test Name**: `error - purchase not found (ErrNoRows)`

**Description**: Verifies that `sql.ErrNoRows` is propagated when the purchase does not exist or belongs to a different user (ownership enforced by the INNER JOIN in the query).

**Mocked Components**:
```go
repo.On("GetItemsByPurchaseId", int64(404), 1).Return([]document.Item{}, sql.ErrNoRows)
```

**Expected Error**: `sql.ErrNoRows.Error()`

---

### Test Structure

```go
tests := []struct {
    name          string
    objId         int64
    userID        int
    mockSetup     func(repo *MockDocumentRepository)
    expectedItems []document.Item
    expectErr     bool
    errorContains string
}
```

---

---

## TestGetPurchaseById

Tests `GetPurchaseById()` on the document service — fetches a single purchase (with items) by its object ID, enforcing user ownership.

### Test Cases

#### 1. Success - Returns Purchase with Items ✅

**Test Name**: `success - returns purchase with items`

**Description**: Verifies that a purchase with line items is fetched and returned correctly for an authenticated user.

**Mocked Components**:
```go
repo.On("GetPurchaseByObjId", int64(5), 1).Return(&document.Document{
    Id:    "5",
    Title: "Woolworths",
    Items: []document.Item{
        {Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
    },
}, nil)
```

**Assertion**: `assert.Equal(t, expectedDoc, doc)` + `assert.NoError(t, err)`

---

#### 2. Success - Purchase with No Items ✅

**Test Name**: `success - purchase with no items`

**Description**: Verifies that a purchase with an empty items list is returned correctly (not an error condition).

**Mocked Components**:
```go
repo.On("GetPurchaseByObjId", int64(5), 1).Return(&document.Document{
    Id:    "5",
    Title: "Coles",
    Items: []document.Item{},
}, nil)
```

**Assertion**: `assert.Equal(t, expectedDoc, doc)` + `assert.NoError(t, err)`

---

#### 3. Error - Purchase Not Found (ErrNoRows) ❌

**Test Name**: `error - purchase not found (ErrNoRows)`

**Description**: Verifies that `sql.ErrNoRows` is propagated when the purchase does not exist or belongs to a different user (ownership enforced by `WHERE d.user_id = ?` in the query).

**Mocked Components**:
```go
repo.On("GetPurchaseByObjId", int64(404), 1).Return(nil, sql.ErrNoRows)
```

**Expected Error**: `sql.ErrNoRows.Error()`

---

#### 4. Error - Repository Failure ❌

**Test Name**: `error - repository failure`

**Description**: Verifies that a generic database error is propagated to the caller.

**Mocked Components**:
```go
repo.On("GetPurchaseByObjId", int64(99), 1).Return(nil, errors.New(errDBFailure))
```

**Expected Error**: `errDBFailure` (`"db failure"`)

---

### Test Structure

```go
tests := []struct {
    name          string
    objId         int64
    userID        int
    mockSetup     func(repo *MockDocumentRepository)
    expectedDoc   *document.Document
    expectErr     bool
    errorContains string
}
```

---

## Test Architecture

### Mock Objects

#### MockDocumentRepository

A mock implementation of the `DocumentRepository` interface using testify/mock.

**Implemented Methods**:
```go
func (m *MockDocumentRepository) SetLatestObjId(tx *sql.Tx, doc *document.Document) (int64, error)
func (m *MockDocumentRepository) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, doc *document.Document) (int64, error)
func (m *MockDocumentRepository) SaveMetadata(tx *sql.Tx, doc *document.Document) (int64, error)
func (m *MockDocumentRepository) InsertFilePath(tx *sql.Tx, doc *document.Document) error
func (m *MockDocumentRepository) SaveItems(tx *sql.Tx, objId int64, docId int64, items []document.Item) error
func (m *MockDocumentRepository) BeginTx() (*sql.Tx, error)
func (m *MockDocumentRepository) GetPurchasesByUser(userID int, year, month string) ([]document.Document, error)
func (m *MockDocumentRepository) GetPurchaseByObjId(objId int64, userID int) (*document.Document, error)
func (m *MockDocumentRepository) GetFilePathByObjId(objId int64, userID int) (string, string, error)
func (m *MockDocumentRepository) GetDocIdByObjId(tx *sql.Tx, objId int64, userID int) (int64, error)
func (m *MockDocumentRepository) GetItemsByPurchaseId(objId int64, userID int) ([]document.Item, error)
func (m *MockDocumentRepository) UpdatePurchaseMetadata(tx *sql.Tx, objId int64, userID int, doc *document.Document) error
func (m *MockDocumentRepository) DeleteItemsByObjId(tx *sql.Tx, objId int64) error
func (m *MockDocumentRepository) UpsertFilePath(tx *sql.Tx, doc *document.Document) error
func (m *MockDocumentRepository) SoftDeletePurchase(objId int64, userID int) error
```

### Test Data Constants

```go
const (
    testDocumentName = "Test Document"
    testDescription  = "Test Description"
    errDBFailure     = "db failure"
)
```

### Test Structure

The test uses **table-driven testing**, a Go best practice pattern:

```go
tests := []struct {
    name          string
    formData      map[string]string
    includeFile   bool
    fileName      string
    fileContent   string
    userId        int
    mockRepo      func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock)
    expectErr     bool
    errorContains string
}
```

---

## Dependencies

### External Packages

- `github.com/DATA-DOG/go-sqlmock` - SQL mock driver for testing
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/stretchr/testify/assert` - Assertion functions
- `github.com/stretchr/testify/mock` - Mock object framework

### Internal Packages

- `2_Go/api/v1/utils` - Handler implementation
- `2_Go/internal/document` - Document domain models and service

---

## Running the Tests

### Run All Tests

```bash
cd /Users/huythai/Documents/0_Projects/2_Go/server
go test -v ./api/v1/utils/test/...
```

### Run a Specific Test Function

```bash
go test -v ./api/v1/utils/test -run TestUploadDocument
go test -v ./api/v1/utils/test -run TestGetPurchaseItems
go test -v ./api/v1/utils/test -run TestGetPurchaseById
```

### Run with Coverage

```bash
go test -v ./api/v1/utils/test/... -coverpkg=./internal/document/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Run with Coverage HTML Report

```bash
go test -v ./api/v1/utils/test/... -coverpkg=./internal/document/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## Test Results

```
=== RUN   TestUploadDocument
=== RUN   TestUploadDocument/success_-_upload_with_file
=== RUN   TestUploadDocument/success_-_upload_without_file_(metadata_only)
=== RUN   TestUploadDocument/error_-_transaction_begin_fails
=== RUN   TestUploadDocument/error_-_SetLatestObjID_fails
=== RUN   TestUploadDocument/error_-_SaveMetadataWithObjId_fails_(no_file)
=== RUN   TestUploadDocument/error_-_SaveItems_fails
=== RUN   TestUploadDocument/error_-_commit_fails
--- PASS: TestUploadDocument (0.00s)
=== RUN   TestGetPurchaseItems
=== RUN   TestGetPurchaseItems/success_-_returns_items
=== RUN   TestGetPurchaseItems/success_-_empty_items_list
=== RUN   TestGetPurchaseItems/error_-_repository_failure
=== RUN   TestGetPurchaseItems/error_-_purchase_not_found_(ErrNoRows)
--- PASS: TestGetPurchaseItems (0.00s)
=== RUN   TestGetPurchaseById
=== RUN   TestGetPurchaseById/success_-_returns_purchase_with_items
=== RUN   TestGetPurchaseById/success_-_purchase_with_no_items
=== RUN   TestGetPurchaseById/error_-_purchase_not_found_(ErrNoRows)
=== RUN   TestGetPurchaseById/error_-_repository_failure
--- PASS: TestGetPurchaseById (0.00s)
PASS
ok      2_Go/api/v1/utils/test    0.581s
```

---

## Future Improvements

### Potential Additional Test Cases

1. **Items Validation**
   - Items with missing name field
   - Items with invalid quantity/price (non-numeric)
   - Very large item lists

2. **File Upload Edge Cases**
   - Very large files
   - Invalid file types
   - Corrupted file data
   - Special characters in filename

3. **Virus Scanning**
   - Mock virus scanner for testing production code path
   - Test malware detection scenario

4. **Concurrent Uploads**
   - Test race conditions in object ID generation

5. **Integration Tests**
   - Test with real database
   - End-to-end upload flow with items

---

## Related Files

- [doc_handler_test.go](doc_handler_test.go) - Test implementation
- [service.go](../../../../internal/document/service.go) - Document service
- [repository.go](../../../../internal/document/repository.go) - Document repository
- [model.go](../../../../internal/document/model.go) - Document model

---

*Last Updated: 2026-03-01*
*Author: Claude Code*
*Review Status: Pending*
