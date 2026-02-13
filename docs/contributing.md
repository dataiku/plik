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

# Go linter (golangci-lint)
make lint

# Backend integration tests (requires Docker)
make test-backends
```

### Running Locally

```bash
cd server
./plikd --config ./plikd.cfg
```

The server starts at [http://127.0.0.1:8080](http://127.0.0.1:8080) and serves both the API and the web interface.

## Code Organization

See the [Architecture Overview](/architecture/system) for details on how the code is structured.
