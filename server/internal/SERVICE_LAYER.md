# Service Layer Responsibilities

**Last Updated:** 2026-02-28

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
| **Create** | `UploadDocument()` (nil file) | Create a purchase without a file attachment (JSON API path) |
| **Update** | `UpdatePurchase()` | Replace metadata and line items of an existing purchase in a transaction |
| **Delete** | `DeletePurchase()` | Soft-delete a purchase (sets `status=-1`) |
| **Read** | `GetPurchases()` | Fetch all active purchases for a user, with optional year/month filter |
| **Read** | `GetPurchaseItems()` | Fetch all line items for a single purchase, verified against the requesting user |
| **Read** | `GetFilePath()` | Fetch the disk path and file name for a user-owned document |
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
- **Read methods are transaction-free** - `GetPurchases`, `GetPurchaseItems`, and `GetFilePath` query directly on `*sql.DB`

### Example Usage

```go
// Upload a document with a file
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

// Create a purchase without a file (pass nil saveFunc — never called when File is nil)
input := &UploadInput{Title: "Coles run", BuyFrom: "Coles", BuyPrice: "42.00", UserId: 1}
err := documentService.UploadDocument(input, nil)

// Update an existing purchase (replaces metadata + line items atomically)
err := documentService.UpdatePurchase(objId, userID, &UploadInput{
    Title: "Updated receipt", BuyFrom: "Woolworths", BuyPrice: "55.00", UserId: userID,
})

// Soft-delete a purchase (sets status = -1)
err := documentService.DeletePurchase(objId, userID)
// errors.Is(err, sql.ErrNoRows) → true when not found or owned by another user

// Fetch purchases (pass "" or "all" to skip a filter)
docs, err := documentService.GetPurchases(userID, "2024", "03")

// Fetch all line items for a specific purchase (ownership enforced in query)
items, err := documentService.GetPurchaseItems(objId, userID)
// returns []Item; empty slice when the purchase has no items
// errors.Is(err, sql.ErrNoRows) → true when not found or owned by another user

// Fetch the file path for a specific document
filePath, fileName, err := documentService.GetFilePath(objId, userID)
```

---

## Document Repository

**Location:** `internal/document/repository.go`

### Interface

| Category | Methods | Description |
|----------|---------|-------------|
| **Write** | `SetLatestObjId()` | Increment and return the next obj_id from `obj_id_counter` (row-locked) |
| **Write** | `SaveMetadataWithObjId()` | Insert a document row into `obj_doc` with a pre-assigned obj_id |
| **Write** | `SaveMetadata()` | Insert a document row without a pre-assigned obj_id |
| **Write** | `InsertFilePath()` | Insert the file path into `obj_doc_path` |
| **Write** | `SaveItems()` | Bulk-insert line items into `obj_item` |
| **Write** | `UpdatePurchaseMetadata()` | `UPDATE obj_doc` for editable fields; returns `sql.ErrNoRows` if not found/owned |
| **Write** | `DeleteItemsByObjId()` | `DELETE FROM obj_item` by `obj_id` (used before re-inserting updated items) |
| **Write** | `SoftDeletePurchase()` | Set `status=-1` on `obj_doc`; returns `sql.ErrNoRows` if not found/owned |
| **Read** | `GetPurchasesByUser()` | Query purchases with optional year/month filter; assembles items per document |
| **Read** | `GetItemsByPurchaseId()` | Fetch line items for a single purchase, verifying user ownership via INNER JOIN |
| **Read** | `GetFilePathByObjId()` | Ownership-checked lookup of file path and name by obj_id |
| **Read** | `GetDocIdByObjId()` | Lookup the auto-increment PK (`id`) of `obj_doc` by `obj_id` (used in update flow) |
| **Tx** | `BeginTx()` | Begin a database transaction |

### Read Query Details

**`GetPurchasesByUser(userID, year, month)`**
- JOINs: `obj_doc` → `obj_item` (LEFT, by `doc_id`) → `obj_doc_path` (LEFT, by `obj_id`)
- Filter: `status = 1` (active only), optional `YEAR(buy_at)` / `MONTH(buy_at)`
- Order: `buy_at DESC, obj_id` (newest first)
- Groups item rows per document in Go using `docMap` + `docOrder` to preserve order

**`GetItemsByPurchaseId(objId, userID)`**
- JOINs: `obj_item` → `obj_doc` (INNER, by `doc_id`)
- Filter: `i.obj_id = ?` (target purchase), `d.user_id = ?` (ownership), `d.status = 1` (active only)
- Returns an empty `[]Item` (not `sql.ErrNoRows`) when a purchase exists but has no items
- The INNER JOIN means items from a deleted or foreign purchase are never returned

**`GetFilePathByObjId(objId, userID)`**
- Verifies ownership (`user_id = ?`) and active status (`status = 1`) in the same query
- Returns empty strings if no file is attached (LEFT JOIN on `obj_doc_path`)
- Returns `sql.ErrNoRows` (wrapped) if the document does not exist or belongs to another user

### Internal Helpers

| Type / Function | Description |
|-----------------|-------------|
| `purchaseRow` struct | Holds raw scanned values from a single JOIN row; avoids long `Scan()` argument lists |
| `purchaseRow.toDocument()` | Converts a scanned row into a `*Document`, parsing the nullable `buy_at` date |
| `purchaseRow.toItem()` | Extracts an `Item` from the row; returns `false` when the LEFT JOIN produced no item |
| `buildPurchaseQuery()` | Builds the base SELECT and appends year/month predicates dynamically |
| `nullableString()` | Converts an empty Go string to `nil` so numeric DB columns receive `NULL` instead of `""` |

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
