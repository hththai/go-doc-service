# Service Layer Responsibilities

**Last Updated:** 2026-02-24

---

## Overview

The service layer acts as the **business logic layer** between HTTP handlers and repositories. Services orchestrate operations, manage transactions, and enforce business rules.

**Dependency Flow:**
```
HTTP Handler → Service → Repository → Database
```

---

## Auth Service

**Location:** `internal/auth/service.go`

### Responsibilities

| Category | Methods | Description |
|----------|---------|-------------|
| **Account Management** | `Register()`, `RegisterAccount()` | Create new user accounts with password hashing |
| **Authentication** | `Login()`, `CheckPassword()` | Validate credentials and authenticate users |
| **Password Operations** | `ChangePassword()`, `ChangePasswordService()` | Handle password updates with validation |
| **Validation** | `ValidateAccountService()`, `ValidateAccountByIdService()` | Verify account credentials |
| **Token Management** | `CreateTokensForUser()`, `RefreshAccessToken()`, `ValidateAccessToken()` | JWT token lifecycle |

### Internal Helpers

| Method | Description |
|--------|-------------|
| `handlePasswordAcctCreation()` | Validate account and hash password for new accounts |
| `handlePasswordUpdate()` | Hash password for updates |
| `hashPassword()` | Generate bcrypt hash with default cost |

### Key Patterns

- **High-level methods** (`RegisterAccount`, `ChangePassword`) manage transactions internally
- **Low-level methods** accept `*sql.Tx` for external transaction control
- **Password hashing** using `bcrypt.DefaultCost`
- **Defined errors:** `ErrInvalidAccount`, `ErrIncorrectPassword`

### Example Usage

```go
// High-level (manages own transaction)
err := authService.RegisterAccount(account)

// Low-level (caller manages transaction)
tx, _ := authRepo.BeginTx()
_, err := authService.Register(tx, account)
tx.Commit()
```

---

## Document Service

**Location:** `internal/document/service.go`

### Responsibilities

| Category | Methods | Description |
|----------|---------|-------------|
| **Document Upload** | `UploadDocument()` | Full upload flow with file handling, metadata, and line items |
| **Transaction** | `BeginTx()` | Start database transactions |
| **Metadata** | `SetLatestObjID()`, `SaveMetadataWithObjId()`, `SaveDocumentMetadata()` | Store document metadata |
| **File Storage** | `SaveFilePath()` | Persist file paths to database |
| **Line Items** | `SaveItems()` | Persist purchase line items to `obj_item` table |

### Internal Helpers

| Function | Description |
|----------|-------------|
| `saveFileAndMetadata()` | Orchestrates temp save, scan, and final storage |
| `buildUploadPath()` | Constructs storage path using folder indexing |
| `saveTemp()` | Saves file to temporary location for processing |
| `fileScan()` | Runs malware detection (production only) |
| `randomString()` | Generates random hex strings for temp folders |

### Key Patterns

- **`FileSaveFunc` injection** - Function parameter for testability
- **`UploadInput` struct** - Decouples service from HTTP layer
- **Automatic cleanup** - Temp files removed via `defer`
- **Environment-aware** - Virus scanning only in production
- **Transaction safety** - Rollback on panic/error

### Example Usage

```go
input := &UploadInput{
    Title:       "Report",
    Description: "Annual report",
    UserId:      123,
    File:        fileHeader,
    Items: []Item{
        {Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
    },
}

err := documentService.UploadDocument(input, func(file *multipart.FileHeader, dst string) error {
    return c.SaveUploadedFile(file, dst) // gin context method
})
```

---

## Service Layer Contract

All services must follow these principles:

### 1. Business Logic Ownership
Services contain all business rules. Handlers only handle HTTP concerns.

### 2. Transaction Management
High-level operations manage their own transactions. Use `BeginTx()` + `defer tx.Rollback()` pattern.

### 3. Validation
Validate input before database operations. Return early on validation failure.

### 4. Error Handling
Return meaningful domain errors. Define error constants for common cases.

### 5. Dependency Injection
Accept repositories via constructor. No direct database access.

```go
func NewAuthService(authRepo AuthRepository) AuthService {
    return &authService{authRepo: authRepo}
}
```

### 6. Testability
- Use interfaces for dependencies
- Accept function parameters for side effects (e.g., `FileSaveFunc`)
- Avoid global state

---

## Adding a New Service

1. **Define interface** in `internal/<module>/service.go`
2. **Implement struct** with repository dependency
3. **Constructor** with `New<Module>Service(repo)`
4. **High-level methods** manage transactions internally
5. **Low-level methods** accept `*sql.Tx` for flexibility
6. **Error constants** for domain-specific errors
7. **Unit tests** with mock repository
