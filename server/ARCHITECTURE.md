# Architecture — Server (`server/`)

> Internals of the Plik HTTP server. For system-wide overview, see the root [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## Package Structure

```
server/
├── main.go         ← entry point (calls cmd.Execute())
├── plikd.cfg       ← default configuration file
├── cmd/            ← CLI commands (cobra)
├── common/         ← shared types, config, feature flags, utilities
├── context/        ← custom request context
├── data/           ← data backend interface + implementations
├── handlers/       ← HTTP handler functions
├── metadata/       ← metadata backend (GORM)
├── middleware/     ← middleware chain components
└── server/         ← HTTP server setup + router
```

---

## `cmd/` — CLI Commands (Cobra)

The server binary `plikd` uses [cobra](https://github.com/spf13/cobra) for CLI management.

| File | Command | Description |
|------|---------|-------------|
| `root.go` | `plikd` | Start the server (default command) |
| `user.go` | `plikd user create/list/delete` | Manage local users |
| `token.go` | `plikd token create/list/delete` | Manage user tokens |
| `file.go` | `plikd file list/delete` | Manage uploads/files |
| `clean.go` | `plikd clean` | Run metadata cleanup |
| `import.go` | `plikd import` | Import metadata from JSON |
| `export.go` | `plikd export` | Export metadata to JSON |

Config loading order: `--config` flag → `PLIKD_CONFIG` env → `./plikd.cfg` → `/etc/plikd.cfg`.

---

## `common/` — Shared Types & Config

Core types used throughout the server:

| File | Content |
|------|---------|
| `upload.go` | `Upload` struct — container for files with TTL, options, password |
| `file.go` | `File` struct + status constants (`missing`/`uploading`/`uploaded`/`removed`/`deleted`) |
| `user.go` | `User` struct + provider constants (`local`/`google`/`ovh`/`oidc`) |
| `token.go` | `Token` struct — UUID-based upload tokens |
| `config.go` | `Configuration` struct — TOML parsing + env var override |
| `feature_flags.go` | Feature flag types: `disabled`/`enabled`/`default`/`forced` |
| `settings.go` | `Setting` struct — server-level key/value (e.g., auth signing key) |
| `authentication.go` | `SessionAuthenticator` — JWT session cookie management |
| `paging.go` | `PagingQuery` — pagination parameters |
| `stats.go` | `ServerStats` — upload/file/user counts |
| `metrics.go` | `PlikMetrics` — Prometheus metric registry |
| `version.go` | Build info (version, git commit, build date) |
| `utils.go` | `GenerateRandomID()`, `StripPrefix()`, etc. |

---

## `context/` — Custom Request Context

> **Historical note**: This package predates Go's stdlib `context.Context` (added in Go 1.7). It provides a typed, mutex-protected struct that carries request-scoped values through the middleware chain.

The `Context` struct holds:

| Field | Type | Set By |
|-------|------|--------|
| `config` | `*Configuration` | Server init |
| `logger` | `*Logger` | Server init |
| `metadataBackend` | `*metadata.Backend` | Server init |
| `dataBackend` | `data.Backend` | Server init |
| `streamBackend` | `data.Backend` | Server init |
| `authenticator` | `*SessionAuthenticator` | Server init |
| `metrics` | `*PlikMetrics` | Server init |
| `sourceIP` | `net.IP` | `SourceIP` middleware |
| `upload` | `*Upload` | `Upload` middleware |
| `file` | `*File` | `File` middleware |
| `user` | `*User` | `Authenticate` middleware |
| `token` | `*Token` | `Authenticate` middleware |
| `pagingQuery` | `*PagingQuery` | `Paginate` middleware |

All fields are accessed via getter/setter methods protected by a `sync.RWMutex`. Getters panic if a required field is nil (fail-fast pattern).

The `context` package also provides `Chain` — a composable middleware chain builder: `NewChain(mw...).Append(mw...).Then(handler)`.

---

## `data/` — Data Backend

The `Backend` interface is minimal (3 methods):

```go
type Backend interface {
    AddFile(file *common.File, reader io.Reader) (err error)
    GetFile(file *common.File) (reader io.ReadCloser, err error)
    RemoveFile(file *common.File) (err error)
}
```

### Implementations

| Package | Backend | Notes |
|---------|---------|-------|
| `data/file` | Local filesystem | Files stored in configurable directory |
| `data/s3` | Amazon S3 / MinIO | Supports SSE-C and S3-managed encryption |
| `data/swift` | OpenStack Swift | |
| `data/gcs` | Google Cloud Storage | |
| `data/stream` | In-memory pipe | Blocks uploader until downloader connects — nothing stored |
| `data/testing` | In-memory map | For tests only |

---

## `metadata/` — Metadata Backend (GORM)

Uses GORM with gormigrate for schema management across SQLite3, PostgreSQL, and MySQL.

### Key behaviors

- **SQLite3**: WAL mode + foreign keys enabled on connect
- **Schema init**: Auto-migrates `Upload`, `File`, `User`, `Token`, `Setting` tables
- **Migrations**: Versioned via gormigrate — see `migrations.go`
- **Cleaning**: `Clean()` removes orphan files and tokens (FK integrity)
- **Metrics**: GORM Prometheus plugin for DB stats

### Files

| File | Purpose |
|------|---------|
| `metadata.go` | Backend init, config, shutdown, clean |
| `migrations.go` | Schema migration definitions |
| `upload.go` | Upload CRUD + listing + expiration |
| `file.go` | File CRUD + status updates |
| `user.go` | User CRUD + listing |
| `token.go` | Token CRUD + listing |
| `setting.go` | Server settings key/value store |
| `stats.go` | Aggregate statistics queries |
| `exporter.go` | JSON export of all data |
| `importer.go` | JSON import |

---

## `middleware/` — Middleware Chain

Each middleware is a function that takes a `context.Context` and optionally calls the next handler.

| File | Middleware | Purpose |
|------|-----------|---------|
| `context.go` | `Context()` | Initialize context with server-level values |
| `log.go` | `Log` | Request/response logging |
| `recover.go` | `Recover` | Panic recovery → HTTP error response |
| `source_ip.go` | `SourceIP` | Extract client IP (supports `X-Forwarded-For` header) |
| `authenticate.go` | `Authenticate(acceptToken)` | Parse session cookie / X-PlikToken header → set user/token |
| `impersonate.go` | `Impersonate` | Admin impersonation support |
| `upload.go` | `Upload` | Resolve `{uploadID}` → load upload + check auth |
| `file.go` | `File` | Resolve `{fileID}` → load file from upload |
| `create_upload.go` | `CreateUpload` | Parse upload creation params for quick upload |
| `paginate.go` | `Paginate` | Parse pagination query params |
| `redirect.go` | `RedirectOnFailure` | Redirect to webapp on error (for browser requests) |
| `user.go` | `User` | Resolve `{userID}` → load user (admin or self) |

---

## `handlers/` — HTTP Handlers

Each handler file contains one or more `http.Handler` functions.

| File | Handlers | Description |
|------|----------|-------------|
| `create_upload.go` | `CreateUpload` | Create upload with options, validate config/quotas |
| `add_file.go` | `AddFile` | Upload file to existing upload (multipart) |
| `get_upload.go` | `GetUpload` | Return upload metadata |
| `get_file.go` | `GetFile` | Download file, handle OneShot, extend TTL |
| `get_archive.go` | `GetArchive` | Download all files as zip |
| `remove_file.go` | `RemoveFile` | Mark file as removed |
| `remove_upload.go` | `RemoveUpload` | Soft-delete upload |
| `misc.go` | `GetConfiguration`, `GetVersion`, `GetQrCode`, `Health` | Utility endpoints |
| `local.go` | `LocalLogin`, `Logout` | Local auth |
| `google.go` | `GoogleLogin`, `GoogleCallback` | Google OAuth |
| `ovh.go` | `OvhLogin`, `OvhCallback` | OVH OAuth |
| `oidc.go` | `OIDCLogin`, `OIDCCallback` | OpenID Connect |
| `me.go` | `UserInfo`, `DeleteAccount`, `GetUserStatistics` | Current user |
| `token.go` | `GetUserTokens`, `CreateToken`, `RevokeToken` | Token management |
| `user.go` | `GetUsers`, `CreateUser`, `UpdateUser` | User management |
| `admin.go` | `GetServerStatistics`, `GetUploads`, `RemoveUserUploads`, `GetUserUploads` | Admin endpoints |

---

## `server/` — HTTP Server Setup

`PlikServer` is the main server struct. It:

1. Initializes backends (metadata, data, stream) and authenticator
2. Builds middleware chains (see root ARCHITECTURE.md for chain table)
3. Configures gorilla/mux router with all routes
4. Starts HTTP server via `net.Listen` + `httpServer.Serve` (supports ephemeral port allocation with `ListenPort: 0`)
5. Starts cleaning routine (if auto-clean enabled)
6. Starts metrics HTTP server (if configured)

After start, call `GetListenPort()` to retrieve the actual listen port (useful when configured with port 0).

Shutdown: graceful with configurable timeout, closes HTTP server + metadata backend.
