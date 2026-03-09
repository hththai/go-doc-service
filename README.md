# Purchase Record App

A full-stack application for managing purchase documents and receipts. Upload, categorize, and review purchase records with OCR-powered data extraction.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25, Gin, Clean Architecture |
| Frontend | React 19, TanStack Router/Query/Form, Tailwind CSS v4 |
| Database | MySQL 8 |
| Cache / Sessions | Redis |
| Proxy | Traefik v3 |
| Containerization | Docker, Docker Compose |

---

## Architecture

```
.
├── server/               # Go backend
│   ├── api/v1/           # HTTP handlers & routes
│   │   ├── authApi/      # Auth endpoints
│   │   ├── categoryApi/  # Category endpoints
│   │   ├── documentApi/  # Document upload/preview endpoints
│   │   └── ocrApi/       # OCR endpoints
│   ├── internal/         # Business logic (Clean Architecture)
│   │   ├── auth/         # Authentication & JWT
│   │   ├── category/     # Category management
│   │   ├── config/       # Centralized configuration
│   │   ├── document/     # Document storage & processing
│   │   ├── log/          # Rotating file logger
│   │   ├── obj/          # Shared domain models
│   │   └── repo/         # Database schema & initialization
│   ├── middleware/       # Rate limiting, JWT auth, request logging
│   └── utils/            # Shared utilities
│
└── client-vite/          # React frontend
    └── client-vite/
        └── src/
            ├── routes/   # File-based routing (TanStack Router)
            ├── api/      # API client functions
            ├── components/
            └── hooks/
```

**Dependency flow:** `HTTP Handler → Service → Repository → Database`

---

## Features

- **Document Management** — Upload, store, and preview purchase receipts (PDF/image)
- **OCR Extraction** — Extract text and data from uploaded documents
- **Category Labels** — Organize documents with user-defined color-coded categories
- **Authentication** — JWT-based auth with access/refresh token rotation
- **Rate Limiting** — Per-IP request throttling via Redis
- **Security** — CORS, file-type validation, virus scanning (LMD/ClamAV), quarantine folder
- **Rotating Logs** — Size-based log rotation with configurable backup count

---

## Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local server development)
- Node.js 22+ with pnpm (for local client development)

---

## Getting Started

### 1. Environment variables

Copy the example env file and fill in values:

```bash
cp server/docker-compose.example.yml docker-compose.override.yml
```

Required variables:

| Variable | Description |
|---|---|
| `DB_ROOT_PASSWORD` | MySQL root password |
| `DB_NAME` | Database name |
| `DB_USER` | Database user |
| `DB_PASSWORD` | Database password |
| `JWT_SECRET` | Secret key for signing JWTs (min. 32 chars) |
| `CORS_ORIGINS` | Allowed origins (comma-separated) |
| `GO_ENV` | `development` or `production` |

### 2. Run with Docker Compose

```bash
docker-compose up --build
```

Services started:
- `api` → `http://localhost:8088`
- `db` → MySQL on port `3307`
- `redis-service` → Redis
- `traefik` → Reverse proxy on ports `80` / `443`

---

## Local Development

### Backend

```bash
cd server
go run main.go
```

Run tests:

```bash
cd server
go test ./...
```

### Frontend

```bash
cd client-vite/client-vite
pnpm install
pnpm dev        # starts on http://localhost:3000
pnpm test       # run Vitest
```

---

## Configuration

All server configuration is loaded once via `internal/config/config.go`:

```go
type Config struct {
    Env      string          // development | production
    Server   ServerConfig    // Port
    Database DatabaseConfig  // Host, Port, User, Password, Name
    JWT      JWTConfig       // Secret
    CORS     CORSConfig      // Allowed origins
    Redis    RedisConfig     // Addr
}
```

Environment files: `server/.env`, `server/.env.development`, `server/.env.production`

---

## API Overview

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/auth/register` | Register a new account |
| `POST` | `/v1/auth/login` | Login and receive tokens |
| `POST` | `/v1/auth/refresh` | Refresh access token |
| `GET` | `/v1/document` | List documents |
| `POST` | `/v1/document/upload` | Upload a purchase document |
| `GET` | `/v1/document/preview/:id` | Preview a document |
| `GET` | `/v1/category` | List categories |
| `POST` | `/v1/category` | Create a category |
| `PUT` | `/v1/category/:guid` | Update a category |
| `DELETE` | `/v1/category/:guid` | Delete a category |

---

## Security Notes

- JWT secrets must be set via environment variable — never hardcoded
- Uploaded files are virus-scanned before storage
- Suspicious files are moved to a quarantine folder
- Rate limiting is enforced in production via Redis
- CORS origins are explicitly allowlisted

---

## License

Private — all rights reserved.
