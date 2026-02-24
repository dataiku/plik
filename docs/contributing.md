# Contributing

Contributions are welcome! Here's how to get involved.

## Getting Help

- **Telegram**: Join the [Plik channel](https://t.me/plik_rootgg) for questions and discussion
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/root-gg/plik/issues)

## Development Setup

### Prerequisites

- Go (see `go.mod` for minimum version)
- Node.js
- Make

### Building

```bash
git clone https://github.com/root-gg/plik.git
cd plik

# Build everything (frontend + server + client)
make

# Build only the server
make server

# Build only the frontend
make frontend

# Build only the client
make client
```

### Running Tests

```bash
# Go unit tests
make test

# Frontend unit tests (vitest)
make test-frontend

# Go linter (golangci-lint)
make lint

# Backend integration tests (requires Docker)
make test-backends
```

### Running Locally

Start the Go backend:

```bash
cd server
./plikd --config ./plikd.cfg
```

The server starts at [http://127.0.0.1:8080](http://127.0.0.1:8080) and serves both the API and the web interface.

#### Webapp dev server

For frontend development with hot-reload, run the Vite dev server which proxies API calls to the Go backend:

```bash
make docs
cd webapp
npm install
npm run dev           # http://localhost:5173
npm run dev -- --host # expose to local network
```

#### Documentation dev server

To preview the documentation site locally:

```bash
cd docs
npm install
npm run dev           # http://localhost:5173/plik/
npm run dev -- --host # expose to local network
```

## AI-Assisted Development

Plik ships with built-in support for AI coding agents (Cursor, Antigravity, Copilot, etc.). If you're using an agentic coding assistant, the repo is pre-configured to give it deep project context.

This is all still very exploratory/experimental, and subject to change as the community is converging to a golden standard. We won't try to adapt to everyone's own IDE/setup.

### Agent Context Files

| File | Purpose |
|------|---------|
| `AGENTS.md` | Entry point for AI agents — tech stack, repo layout, build/test commands, conventions |
| `ARCHITECTURE.md` | System-wide architecture overview |
| `server/ARCHITECTURE.md` | Server internals (handlers, middleware, backends) |
| `client/ARCHITECTURE.md` | CLI client (commands, config, crypto, archive, MCP) |
| `webapp/ARCHITECTURE.md` | Vue 3 SPA (components, routing, state) |
| `plik/ARCHITECTURE.md` | Go client library (public API, types) |

Your agent should read `AGENTS.md` first, then follow pointers to scoped `ARCHITECTURE.md` files for the area being worked on.

### Agentic Workflows

Pre-built workflows live in `.agents/workflows/` and can be invoked as slash commands:

| Command | What it does |
|---------|-------------|
| `/review-changes` | Critical self-review of local changes (lint, build, test, code review checklist) |
| `/prepare-pr` | Full PR preparation pipeline (review → commit → push → draft PR) |

## Code Organization

See the [Architecture Overview](/architecture/system) for details on how the code is structured.
