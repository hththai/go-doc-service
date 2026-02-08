# Server Architecture Review & TODO

**Date:** 2026-02-01
**Last Updated:** 2026-02-08 (configuration centralized)
**Total Lines of Code:** ~4,234 lines of Go code
**Test Files:** 6 test files

---

## Current Architecture Overview

### Structure
```
server/
├── api/v1/           # HTTP Layer (handlers, routes, utils)
│   ├── authApi/
│   ├── documentApi/
│   ├── routers/
│   └── utils/
├── internal/         # Business Logic Layer
│   ├── auth/         # Authentication & authorization
│   ├── config/       # Centralized configuration ✅
│   ├── document/     # Document management
│   ├── log/          # Logging models
│   ├── obj/          # Shared domain objects
│   └── repo/         # Database schema & initialization
├── middleware/       # Middleware components
│   └── authen/       # JWT authentication
└── utils/            # Shared utilities
```

### Pattern
- **Architecture:** Layered/Clean Architecture
- **Core Pattern:** Repository + Service
- **Dependency Flow:** HTTP → Service → Repository → Database

---

## Architecture Strengths ✅

1. **Clear Layering** - Well-defined HTTP, Service, Repository layers
2. **Interface-Based Design** - Repositories use interfaces (testable with mocks)
3. **API Versioning** - Clean `/v1/` structure allows API evolution
4. **Middleware Pattern** - Auth and rate limiting properly separated
5. **Test Infrastructure** - Mock-based testing framework in place
6. **Transaction Safety** - Defer rollback pattern consistently used
7. **Environment Separation** - Development vs production configs
8. **Containerization** - Docker and docker-compose ready

---

## Critical Issues

### 1. Business Logic in Wrong Layer ✅ COMPLETED

**Previous Problem:**
- `api/v1/utils/register_handler.go` (417 lines) - ⚠️ Auth logic partially extracted to auth service
- `api/v1/utils/document_handler.go` (251 lines) - ❌ Still contains file upload business logic

**Current State:**
- ✅ Auth logic moved to `internal/auth/service.go`
- ✅ File handling moved to `internal/document/service.go`
- ✅ `document_handler.go` deleted (no longer needed)
- ✅ `register_handler.go` deleted (no longer needed)

**Completed:**
- ✅ Extract business logic from `api/v1/utils/` to appropriate service layers
- ✅ Keep utils package for actual utilities (validation, sanitization, helpers)
- ✅ Move file handling logic from `api/v1/utils/document_handler.go` to `internal/document/service.go`
- ✅ Move authentication orchestration to `internal/auth/service.go` (RegisterAccount, ChangePassword, token operations)

---

### 2. Handlers Accessing Database Directly ✅ COMPLETED

**Previous Problem:**
```go
type AuthHandler struct {
    AccountSvc auth.AuthService
    DB         *sql.DB  // ❌ Should not directly access DB
    Logger     logrus.FieldLogger
}
```

**Current State:**
```go
type AuthHandler struct {
    AccountSvc auth.AuthService
    Logger     logrus.FieldLogger
}  // ✅ Clean - no DB dependency
```

**Files Updated:**
- `/server/api/v1/authApi/handler.go` ✅
- `/server/api/v1/documentApi/handler.go` ✅

**Completed:**
- ✅ Remove `*sql.DB` from all handler structs
- ✅ Handlers should only depend on services
- ✅ Services manage all database transactions
- ✅ Update handler initialization in `main.go`
- ✅ Repositories expose `BeginTx()` for transaction support

---

### 3. Scattered Configuration ✅ COMPLETED

**Previous Problem:**
- Config loading in multiple places:
  - `internal/repo/config.go` - DB config
  - `utils/utils.go` - JWT config loading
  - `main.go` - Environment detection
- No centralized configuration struct

**Current State:**
- ✅ Created `internal/config/config.go` with centralized Config struct
- ✅ All configuration loaded once via `config.Load()`
- ✅ Separated database schema to `internal/repo/schema.go`
- ✅ Created `internal/repo/database.go` for DB initialization using config
- ✅ Updated `middleware/authen/` to use `config.Get().JWT.Secret`
- ✅ Updated `utils/utils.go` to delegate `IsProduction()` to config
- ✅ Updated `main.go` to use centralized config

**Files Created/Updated:**
- `/server/internal/config/config.go` - NEW: Centralized config with Config struct
- `/server/internal/repo/schema.go` - NEW: Database migrations
- `/server/internal/repo/database.go` - NEW: InitDB using config
- `/server/internal/repo/config.go` - DELETED (split into schema.go + database.go)
- `/server/middleware/authen/authenWithJWT.go` - Uses config.Get().JWT.Secret
- `/server/utils/utils.go` - Delegates to config.IsProduction()
- `/server/main.go` - Uses config.Load() and passes config to services

**Config Structure:**
```go
type Config struct {
    Env      string
    Server   ServerConfig   // Port
    Database DatabaseConfig // User, Password, Host, Port, Name
    JWT      JWTConfig      // Secret
    CORS     CORSConfig     // Origins
    Redis    RedisConfig    // Addr
}
```

---

### 4. Inconsistent Service Layer ✅ COMPLETED

**Previous Problem:**
- `document` module had thin service layer (just delegates to repository)

**Current State:**
- `auth` module: ✅ Full repository+service pattern with business logic and internal tx management
- `document` module: ✅ Full service pattern with UploadDocument, file handling, tx management

**Files Updated:**
- `/server/internal/auth/service.go` - ✅ Full service (RegisterAccount, ChangePassword, token ops)
- `/server/internal/document/service.go` - ✅ Full service (UploadDocument, file handling, virus scan)

**Completed:**
- ✅ Add proper business logic to `internal/document/service.go` (moved from api/v1/utils/document_handler.go)
- ✅ Move file validation logic from handlers to document service
- ✅ Standardize service method signatures across modules
- ✅ Document service layer responsibilities (see `internal/SERVICE_LAYER.md`)

---

### 5. Missing DTO Layer ❌

**Problem:**
- Using domain models directly in API requests/responses
- No separation between API contracts and domain models
- Changes to domain affect API contracts

**Fix:**
- [ ] Create `api/v1/models/` package for DTOs
- [ ] Define request models (RegisterRequest, LoginRequest, etc.)
- [ ] Define response models (AuthResponse, DocumentResponse, etc.)
- [ ] Keep domain models in `internal/*/model.go`
- [ ] Add mapping functions between DTOs and domain models

---

## Medium Priority Issues

### 6. Transaction Management Scattered ⚠️ (Partial Progress)

**Previous Problem:**
- Transaction handling in `api/v1/utils/` handlers
- Inconsistent transaction patterns

**Current State:**
- ✅ Transaction handling moved to service layer (auth and document services)
- ⚠️ Services still expose `*sql.Tx` in some method signatures

**Fix:**
- [ ] Implement Unit of Work pattern
- ✅ Centralize transaction management in service layer
- [ ] Service methods should not expose `*sql.Tx` parameters
- [ ] Consider using a transaction middleware or context

---

### 7. Missing Abstractions

**Fix:**
- [ ] Add centralized error handling
  - Custom error types for domain errors
  - Consistent error response format
  - Error logging strategy
- [ ] Add request validation layer
  - Validation middleware before handlers
  - Separate from domain validation
- [ ] Add response builders/formatters

---

### 8. Incomplete Modules

**Files with Issues:**
- `/server/internal/log/model.go` - Only has empty struct
- `/server/internal/repo/services.go` - Unused interface
- Multiple TODO comments indicating incomplete work

**Fix:**
- [ ] Complete or remove `internal/log/` module
- [ ] Remove unused code from `internal/repo/services.go`
- [ ] Address all TODO comments in codebase
- [ ] Clean up `ValidateAccountService` (marked for deletion but still in use)

---

### 9. Security Concerns

**Issues:**
- JWT secret loaded from environment but no validation
- File upload validation minimal
- Rate limiting only in production mode
- Hardcoded token in previewPDF: `token != "abc123"`

**Fix:**
- [ ] Validate JWT secret exists and meets minimum requirements
- [ ] Add comprehensive file upload validation
  - File type validation
  - File size limits
  - Virus scanning (if needed)
- [ ] Enable rate limiting in all environments
- [ ] Remove hardcoded tokens, use proper authentication
- [ ] Add input sanitization for all user inputs

---

### 10. Testing Gaps

**Current Test Coverage:**
- ✅ `internal/auth/test/auth_test.go`
- ✅ `api/v1/authApi/authApi_test/handler_test.go`
- ✅ `api/v1/utils/test/register_handler_test.go`
- ✅ `api/v1/utils/test/doc_handler_test.go`
- ✅ `middleware/authen/authen_test/authenWithJWT_test.go`
- ✅ `utils/test/utils_test.go`
- ❌ No tests for `internal/document/` module
- ❌ No integration tests
- ❌ No API-level end-to-end tests

**Fix:**
- [ ] Add unit tests for `internal/document/service.go`
- [ ] Add unit tests for `internal/document/repository.go`
- [ ] Add integration tests for database operations
- [ ] Add API contract tests (e2e tests)
- [ ] Add test for rate limiting middleware
- [ ] Aim for >80% code coverage

---

### 11. Code Organization

**Issues:**
- Some files too long (register_handler.go: 417 lines)
- Commented-out code blocks should be removed
- Hardcoded values (INDEX_FOLDER = 100, file paths)

**Fix:**
- [ ] Break down large files into smaller, focused files
- [ ] Remove commented-out code (use git history if needed)
- [ ] Move hardcoded values to configuration
- [ ] Extract constants to dedicated files

---

### 12. Documentation

**Fix:**
- [ ] Add architecture documentation (README.md)
- [ ] Document API endpoints (consider Swagger/OpenAPI)
- [ ] Add inline comments for complex business logic
- [ ] Document configuration options
- [ ] Add setup/deployment guides

---

## Long Term Improvements

### 13. Dependency Injection

**Current State:**
- Manual DI via Dependencies struct
- Handler initialization in `main.go` (232 lines)

**Fix:**
- [ ] Consider DI framework (google/wire, uber/fx)
- [ ] Auto-generate dependency wiring
- [ ] Simplify `main.go`

---

### 14. Observability

**Fix:**
- [ ] Implement structured logging throughout
  - Consistent log format
  - Request ID tracking
  - Log levels properly used
- [ ] Add metrics collection (Prometheus)
- [ ] Add distributed tracing (Jaeger/OpenTelemetry)
- [ ] Add health check endpoints

---

### 15. Advanced Patterns (If Needed)

**Consider when complexity grows:**
- [ ] Domain-Driven Design (DDD)
  - Aggregate roots
  - Value objects
  - Domain events
- [ ] CQRS for complex queries
  - Separate read/write models
  - Event sourcing (if needed)
- [ ] API Gateway pattern (if moving to microservices)

---

## Action Plan

### Phase 1: Immediate (High Priority) - Weeks 1-2

1. **Extract Business Logic from Utils** ✅
   - ✅ Move file handling from `api/v1/utils/document_handler.go` to `internal/document/service.go`
   - ✅ Move auth logic from `api/v1/utils/register_handler.go` to `internal/auth/service.go`
   - ✅ Token operations (CreateTokensForUser, RefreshAccessToken, ValidateAccessToken) moved to auth service
   - ✅ High-level operations (RegisterAccount, ChangePassword) with internal transaction management
   - ✅ Document service now handles: UploadDocument, file temp storage, virus scanning, path building

2. **Remove DB Dependencies from Handlers** ✅
   - ✅ Update `AuthHandler` struct - remove `DB` field
   - ✅ Update `DocumentHandler` struct - remove `DB` field
   - ✅ Services handle all DB operations
   - ✅ Update `main.go` handler initialization
   - ✅ Repositories now expose `BeginTx()` for transaction support

3. **Centralize Configuration** ✅
   - ✅ Create `internal/config/config.go`
   - ✅ Define `Config` struct
   - ✅ Migrate all config loading to config package
   - ✅ Update all packages to use centralized config

4. **Add DTO Layer**
   - [ ] Create `api/v1/models/` directory
   - [ ] Define request DTOs (auth, document)
   - [ ] Define response DTOs
   - [ ] Add DTO↔Domain mapping functions

5. **Standardize Service Layer** ✅
   - ✅ Enhance `internal/document/service.go` with business logic
   - ✅ Ensure consistent service patterns
   - [ ] Move validation to services

### Phase 2: Medium Term - Weeks 3-4

6. **Unit of Work Pattern**
   - [ ] Design transaction management pattern
   - [ ] Implement UoW interface
   - [ ] Refactor services to use UoW

7. **Validation Layer**
   - [ ] Create validation middleware
   - [ ] Add request validators
   - [ ] Separate API validation from domain validation

8. **Error Handling**
   - [ ] Define custom error types
   - [ ] Create error response builder
   - [ ] Standardize error logging

9. **Complete Test Coverage**
   - [ ] Add document module tests
   - [ ] Add integration tests
   - [ ] Add API contract tests
   - [ ] Set up CI/CD with test coverage reports

10. **Security Hardening**
    - [ ] Validate JWT secret
    - [ ] Enhance file upload validation
    - [ ] Enable rate limiting everywhere
    - [ ] Remove hardcoded credentials

### Phase 3: Long Term - Weeks 5+

11. **Clean Up Codebase**
    - [ ] Remove unused code
    - [ ] Address all TODOs
    - [ ] Break down large files
    - [ ] Extract constants

12. **Documentation**
    - [ ] Architecture documentation
    - [ ] API documentation (Swagger)
    - [ ] Setup guides
    - [ ] Code comments

13. **Observability**
    - [ ] Structured logging
    - [ ] Metrics collection
    - [ ] Distributed tracing
    - [ ] Health checks

14. **Dependency Injection**
    - [ ] Evaluate DI frameworks
    - [ ] Implement chosen solution
    - [ ] Simplify main.go

15. **Advanced Patterns (As Needed)**
    - [ ] Evaluate DDD patterns
    - [ ] Consider CQRS
    - [ ] Plan for scaling

---

## Key Files Reference

### Critical Files to Refactor
- `/server/main.go` - needs simplification

### Completed Refactoring ✅
- `/server/api/v1/authApi/handler.go` - Clean handler, no DB dependency
- `/server/api/v1/documentApi/handler.go` - Clean handler, no DB dependency, uses service for uploads
- `/server/internal/auth/service.go` - Full service with transaction management
- `/server/internal/document/service.go` - Full service with UploadDocument, file handling, tx management
- `/server/internal/config/config.go` - Centralized configuration with Config struct
- `/server/internal/repo/schema.go` - Database migrations (extracted from old config.go)
- `/server/internal/repo/database.go` - InitDB using centralized config
- `/server/api/v1/utils/document_handler.go` - DELETED (logic moved to document service)
- `/server/api/v1/utils/register_handler.go` - DELETED (logic moved to auth service)
- `/server/internal/repo/config.go` - DELETED (split into schema.go + database.go)

### Good Examples to Follow
- `/server/internal/auth/service.go` - Good service layer pattern with internal tx management
- `/server/internal/auth/repository.go` - Good repository pattern with BeginTx()
- `/server/internal/document/service.go` - Good service pattern with FileSaveFunc injection for testability
- `/server/internal/auth/test/auth_test.go` - Good test patterns

### Configuration Files
- `/server/.env` - Base environment variables
- `/server/.env.development` - Dev config
- `/server/.env.production` - Prod config
- `/server/docker-compose.yml` - Container setup

---

## Notes

- **Pattern:** Repository + Service is appropriate for this application
- **Foundation:** Solid, just needs refinement
- **Priority:** Focus on separation of concerns and reducing coupling
- **Goal:** More maintainable, testable, and scalable codebase

**Remember:** Don't over-engineer. Make changes incrementally and test thoroughly.
