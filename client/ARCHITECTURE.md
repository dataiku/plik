# Architecture — CLI Client (`client/`)

> The Plik command-line client. For system-wide overview, see the root [ARCHITECTURE.md](../ARCHITECTURE.md).

---

## Structure

```
client/
├── plik.go        ← main CLI logic (docopt-based argument parsing + upload flow)
├── mcp.go         ← MCP (Model Context Protocol) server over stdio for AI assistants
├── config.go      ← configuration loading (.plikrc)
├── login.go       ← CLI device auth flow (--login)
├── progress.go    ← upload progress bar
├── update.go      ← self-update mechanism
├── archive/       ← archive backends (tar, zip)
├── crypto/        ← crypto backends (openssl, pgp)
├── .plikrc        ← example client configuration
├── plik.sh        ← bash upload wrapper
├── test.sh        ← CLI integration tests
├── test_downgrade.sh  ← client version downgrade tests
└── test_upgrade.sh    ← client version upgrade tests
```

---

## Key Components

### CLI Entry Point (`plik.go`)

Uses [docopt-go](https://github.com/docopt/docopt-go) for argument parsing. The main flow:

1. Parse CLI args (file paths, options like `--oneshot`, `--stream`, `--ttl`, etc.)
2. Load config from `.plikrc` (or `PLIKRC` env var)
3. Handle special modes: `--mcp` (MCP server), `--login` (device auth), `--update` (self-update), `--version`
4. Create upload via the Go library (`plik/`)
5. Add files (with optional archive/encrypt preprocessing)
6. Upload files with progress bars
7. Output results:
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

Archives wrap multiple files/directories into a single upload file.

### Crypto Backends (`crypto/`)

| Backend | Description |
|---------|-------------|
| `openssl` | Symmetric encryption via OpenSSL CLI (configurable cipher) |
| `pgp` | Asymmetric encryption via GPG/PGP (recipient-based) |
| `age` | Modern encryption via [age](https://age-encryption.org/). Supports passphrase, X25519, SSH recipients (`@github_user`, URL, raw key), and SSH host key scanning (`ssh://hostname`). **Default backend.** Sets `upload.E2EE = "age"` for webapp interop (passphrase mode only) |

Encryption wraps the file data stream before upload.

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

## Integration Tests

- `test.sh` — comprehensive CLI integration tests (requires a running server)
- `test_upgrade.sh` / `test_downgrade.sh` — version compatibility tests (stale, unused since ~2021)

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
