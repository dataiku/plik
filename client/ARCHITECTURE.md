# Architecture — CLI Client (`client/`)

> The Plik command-line client. For system-wide overview, see the root [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## Structure

```
client/
├── plik.go          ← entry point: arg parsing, config loading, dispatch
├── app.go           ← PlikCLI struct: upload flow, helpers (Run, info, getFileCommand, printf)
├── mcp.go           ← MCP (Model Context Protocol) server over stdio for AI assistants
├── config.go        ← configuration loading (.plikrc)
├── config_test.go   ← unit tests for config parsing (TTL, password, flags, file loading)
├── login.go         ← CLI device auth flow (--login)
├── progress.go      ← upload progress bar
├── update.go        ← self-update mechanism (PlikCLI method)
├── update_test.go   ← unit tests for update flow (early exits, error handling)
├── archive/         ← archive backends (tar, zip) — errors via CloseWithError
├── crypto/          ← crypto backends (openssl, pgp, age) — errors via CloseWithError
├── .plikrc          ← example client configuration
├── plik.sh          ← bash upload wrapper
└── test.sh          ← CLI integration tests
```

---

## Key Components

### CLI Entry Point (`plik.go`) and Runtime State (`app.go`)

`plik.go` is a slim `main()` using [docopt-go](https://github.com/docopt/docopt-go) for argument parsing. It delegates all upload logic to the `PlikCLI` struct defined in `app.go`.

**`PlikCLI` struct** encapsulates all mutable runtime state:
- `Config`, `Arguments` — parsed configuration and CLI args
- `ArchiveBackend`, `CryptoBackend` — initialized lazily during `Run()`

**`main()` flow** (in `plik.go`):
1. Parse CLI args → early exits: `--version`, `--mcp`, `--info`, `--login`
2. Load config from `.plikrc` → `NewPlikCLI(config, args)`
   - First-run wizard is skipped when `--quiet`, `--yes`, or `--server` is set
3. Dispatch to `cli.Run(client)` for the upload flow

**`PlikCLI.Run()` flow** (in `app.go`):
1. Create upload via the Go library (`plik/`)
2. Add files (with optional archive/encrypt preprocessing)
3. Upload files with progress bars
4. Output results:
   - Default: print download URLs/commands to stdout
   - `--quiet`: print only file URLs to stdout
   - `--json`: print `UploadWithURL` as pretty-printed JSON to stdout (implies `--quiet`)

### Configuration (`config.go`)

Config is a TOML file loaded from (in order):
1. `PLIKRC` environment variable
2. `~/.plikrc`
3. `/etc/plik/plikrc`

Key config fields: `URL` (server), `Token` (user authentication token), archive/crypto defaults.

### CLI Login (`login.go`)

Implements a device authorization flow for CLI authentication:
1. POST `/auth/cli/init` with hostname → receives a code, secret, and verification URL
2. Opens verification URL in user's browser (best-effort)
3. Polls POST `/auth/cli/poll` with code + secret every 2s
4. On approval, saves the token to `~/.plikrc` and exits

Triggered by `--login` flag or interactively during first-run when auth is enabled/forced.

### Archive Backends (`archive/`)

| Backend | Description |
|---------|-------------|
| `tar` | Create tar archives with compression (gzip, bzip2, xz, lzip, lzma, lzop) |
| `zip` | Create zip archives |

Archives wrap multiple files/directories into a single upload file. Errors are propagated via `io.PipeWriter.CloseWithError()` from the archiving goroutine.

### Crypto Backends (`crypto/`)

| Backend | Description |
|---------|-------------|
| `openssl` | Symmetric encryption via OpenSSL CLI (configurable cipher) |
| `pgp` | Asymmetric encryption via GPG/PGP (recipient-based) |
| `age` | Modern encryption via [age](https://age-encryption.org/). Supports passphrase, X25519, SSH recipients (`@github_user`, URL, raw key), and SSH host key scanning (`ssh://hostname`). URLs can serve SSH keys **and** native `age1…` recipients. Plain HTTP URLs trigger a MITM security prompt (default: decline). **Default backend.** Sets `upload.E2EE = "age"` for webapp interop (passphrase mode only) |

Encryption wraps the file data stream before upload. Errors are propagated via `io.PipeWriter.CloseWithError()` from the encryption goroutine.

When the `age` backend is used, the upload is flagged as E2EE (`upload.E2EE = "age"`). This tells the webapp to prompt for a passphrase on download and decrypt client-side. A cryptographically-secure passphrase is auto-generated when none is provided.

### Self-Update (`update.go`)

The client can update itself by downloading the latest matching binary from the configured Plik server. It compares versions and replaces the current binary in-place.

### MCP Server (`mcp.go`)

Implements a local [Model Context Protocol](https://modelcontextprotocol.io/) server over stdio, enabling AI coding assistants (Cursor, VS Code Copilot, etc.) to upload files via Plik. Activated by `plik --mcp`.

Uses the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) (`mcp.StdioTransport`) and the `plik/` Go library for uploads.

**Tools:**
| Tool | Description |
|------|-------------|
| `upload_text` | Upload inline text content as a named file |
| `upload_file` | Upload a single file by path |
| `upload_files` | Upload multiple files by paths in a single upload |
| `server_info` | Get server version, config, and capabilities |

**Prompts:** `upload_guide`

All upload tools use `plik.UploadParams` via struct embedding and return `UploadWithURL` — the standard `common.Upload` metadata enriched with computed URLs.

---

## Tests

### Unit Tests
- `config_test.go` — TTL parsing, password splitting, boolean flags, config file loading, defaults
- `update_test.go` — auto-update disabled, quiet mode, unreachable server, missing platform binary
- `crypto/age/age_test.go` — recipient resolution, encryption round-trips

### Integration Tests
- `test.sh` — comprehensive CLI integration tests (requires a running server)

---

## Conventions

### Stderr for all non-data output

Because `--quiet` and `--json` modes reserve stdout exclusively for machine-readable data (file URLs or JSON), **all** informational, diagnostic, and error messages in the CLI must be written to **stderr** (`fmt.Fprintf(os.Stderr, ...)`). This includes:

- Passphrase display (crypto backends)
- Recipient resolution progress (age backend)
- Debug output
- Streaming download commands
- Archive/crypto error messages
- Progress bars (already write to stderr via the `pb` library)

Never use `fmt.Printf` / `fmt.Println` for non-data output in the CLI.
