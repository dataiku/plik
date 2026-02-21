# AGENTS.md — Plik

> Entry point for AI agents working on this codebase. Not an exhaustive manual — follow pointers to scoped ARCHITECTURE.md files for deeper context.

## What is Plik?

Plik is a temporary file upload system (WeTransfer-like) written in Go, with a Vue 3 web UI and a cross-platform CLI client. It supports multiple storage and metadata backends, authentication providers, and features like one-shot downloads, streaming, end-to-end encryption (E2EE via age), and server-side encryption.

## Tech Stack

| Layer     | Tech |
|-----------|------|
| Server    | Go, gorilla/mux, GORM |
| Webapp    | Vue 3, Vite, Tailwind CSS, CodeMirror 6 |
| CLI       | Go, docopt-go |
| Config    | TOML (server), TOML (client `.plikrc`) |
| Data      | File, OpenStack Swift, S3, Google Cloud Storage |
| Metadata  | SQLite3, PostgreSQL, MySQL (via GORM) |
| CI        | GitHub Actions (tests, docker build/deploy on PR comment, release, Helm chart publish) |

## Repo Layout

```
plik/
├── AGENTS.md              ← you are here
├── ARCHITECTURE.md         ← system-wide architecture
├── README.md               ← project README (concise)
├── Makefile                ← build orchestration
├── Dockerfile
├── .agents/                ← agentic workflows (/review-changes, /prepare-pr)
├── server/                 ← Go server (see server/ARCHITECTURE.md)
│   ├── main.go             ← entry point
│   ├── plikd.cfg           ← default config
│   ├── cmd/                ← CLI commands (cobra)
│   ├── common/             ← shared types, config, feature flags
│   ├── context/            ← custom request context (predates Go stdlib context)
│   ├── data/               ← data backend interface + implementations
│   ├── handlers/           ← HTTP handlers
│   ├── metadata/           ← metadata backend (GORM)
│   ├── middleware/          ← middleware chain (auth, logging, upload/file resolution)
│   └── server/             ← HTTP server + router setup
├── client/                 ← CLI client + MCP server (see client/ARCHITECTURE.md)
├── plik/                   ← Go client library (see plik/ARCHITECTURE.md)
├── webapp/                 ← Vue 3 SPA (see webapp/ARCHITECTURE.md)
├── testing/                ← backend integration tests (see testing/ARCHITECTURE.md)
├── charts/                 ← Helm chart for Kubernetes deployment
├── .github/                ← GitHub Actions workflows (see .github/ARCHITECTURE.md)
├── changelog/              ← release changelogs
├── releaser/               ← release build scripts
├── docs/                   ← VitePress documentation site
└── vendor/                 ← Go vendored dependencies
```

## Build & Run

```bash
make                        # Build everything (frontend + clients + server)
make server                 # Build server only → server/plikd
make client                 # Build CLI client only → client/plik
make frontend               # Build Vue webapp → webapp/dist
make docker                 # Build Docker image (rootgg/plik:dev)
cd server && ./plikd        # Run server on http://127.0.0.1:8080

#### Pull Request Deployments (GitHub Actions)
- `docker build` (comment on PR): Builds and pushes image `rootgg/plik:pr-{PR_NUMBER}`
- `docker deploy` (comment on PR): Deploy PR image to `plik.root.gg` (requires secrets)
```

## Test

```bash
make test                   # Unit tests + CLI integration tests
make test-backends           # Docker-based backend integration tests (all)
make test-backend mariadb    # Docker-based test for a single backend
make lint                   # go fmt + go vet + go fix
make gofix                  # Run go fix
make vuln                   # govulncheck (report only)
```

## Key Files

| File | Purpose |
|------|---------|
| `server/plikd.cfg` | Server configuration (TOML) — all options with comments |
| `client/.plikrc` | CLI client configuration template |
| `Makefile` | Build targets for server, client, frontend, docker, release |
| `server/common/config.go` | Config struct + parsing + env var override logic |
| `server/common/file.go` | File model + status constants |
| `server/common/upload.go` | Upload model |
| `server/common/feature_flags.go` | Feature flag types (`disabled`/`enabled`/`default`/`forced`) |

## Conventions

- **Configuration**: TOML file + env var override using SCREAMING_SNAKE_CASE (e.g., `PLIKD_DEBUG_REQUESTS=true`)
- **Feature flags**: Four states — `disabled`, `enabled` (opt-in), `default` (opt-out), `forced`
- **Special values**: `0` = use server default, `-1` = unlimited (for file size, TTL, etc.)
- **Error handling**: Handlers return HTTP errors; middleware chain panics on missing required context values
- **ID generation**: Random hex strings (16 chars for files, 16 chars for uploads)
- **Backend interface**: `data.Backend` is the storage abstraction; implementations are swappable via config

## Best Practices

- **Always update docs**: When changing code, update the relevant `ARCHITECTURE.md` and VitePress docs
- **Keep Helm chart in sync with plikd config**: When adding, removing, or renaming configuration fields in `server/common/config.go` or `server/plikd.cfg`, you **must** also update the Helm chart:
  - `charts/plik/values.yaml` — add/update the field under `plikd:`
  - `charts/plik/templates/configmap.yaml` — add/update the explicit key in the template
  - `charts/plik/templates/secret.yaml` — if the field is sensitive, add env var injection
- **Helm persistence**: the chart has two independent PVCs — `persistence` for uploaded files (`/home/plik/server/files`) and `dbPersistence` for the SQLite database (`/home/plik/server/db`). Both default to `emptyDir` when disabled. The default `MetadataBackendConfig.ConnectionString` is `/home/plik/server/db/plik.db`.
- **Run tests before committing**: `make lint && make test`
- **Keep ARCHITECTURE.md files in sync**: Each root folder has its own — update the one closest to your change

## Documentation

The documentation lives in two places:

1. **For agents**: Scoped `ARCHITECTURE.md` files in each root folder
2. **For humans**: VitePress site in `docs/` — preview locally with `cd docs && npm run dev`

### Updating docs

```bash
cd docs && npm install       # First time only
cd docs && npm run dev       # Preview at localhost:5173
make docs                    # Build docs (validates links, injects version)
```

**Important**: Always run `make docs` when you touch documentation files to catch build errors (dead links, etc.) before committing.

## Scoped Architecture Docs

| File | Scope |
|------|-------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System-wide: package layering, data flow, API, auth, config |
| [server/ARCHITECTURE.md](server/ARCHITECTURE.md) | Server internals: packages, middleware chain, handlers |
| [client/ARCHITECTURE.md](client/ARCHITECTURE.md) | CLI client: commands, config, archive/crypto |
| [plik/ARCHITECTURE.md](plik/ARCHITECTURE.md) | Go library: public API, types, test harness |
| [webapp/ARCHITECTURE.md](webapp/ARCHITECTURE.md) | Vue 3 SPA: components, routing, API layer, state |
| [testing/ARCHITECTURE.md](testing/ARCHITECTURE.md) | Backend integration tests: docker-based test scripts |
| [releaser/ARCHITECTURE.md](releaser/ARCHITECTURE.md) | Release tooling: build pipeline, Docker stages, client/server compilation |
| [.github/ARCHITECTURE.md](.github/ARCHITECTURE.md) | GitHub Actions workflows, CI/CD, Helm chart release flow |
