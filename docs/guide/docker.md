# Docker Deployment

Plik provides multiarch Docker images for production deployments.

## Quick Start

```bash
docker run -p 8080:8080 rootgg/plik
```

Available tags:
- `latest` — Latest stable release
- `1.3.8` — Specific version
- `dev` — Latest build from master branch

## Custom Configuration

Mount a custom `plikd.cfg` file:

```bash
docker run -p 8080:8080 \
  -v /path/to/plikd.cfg:/home/plik/server/plikd.cfg \
  rootgg/plik
```

See the [Configuration Guide](./configuration.md) for all available options.

## Persistent Storage

Mount a data directory for file storage:

```bash
docker run -p 8080:8080 \
  -v /data:/home/plik/server/files \
  rootgg/plik
```

## Docker Compose

Create a `docker-compose.yml`:

```yaml
version: "2"
services:
  plik:
    image: rootgg/plik:latest
    container_name: plik
    volumes:
      - ./plikd.cfg:/home/plik/server/plikd.cfg
      - ./data:/data
    ports:
      - 8080:8080
    restart: "unless-stopped"
```

Configure `plikd.cfg` to use the mounted volume:

```toml
DataBackend = "file"
[DataBackendConfig]
    Directory = "/data/files"

[MetadataBackendConfig]
    Driver = "sqlite3"
    ConnectionString = "/data/plik.db"
```

Start the container:

```bash
docker-compose up -d
```

## Environment Variables

You can override config values via environment variables:

```bash
docker run -p 8080:8080 \
  -e PLIKD_MAX_FILE_SIZE_STR=10GB \
  -e PLIKD_DEFAULT_TTL_STR=86400 \
  rootgg/plik
```

## Health Check

The Docker image includes a health check endpoint at `/health`:

```bash
curl http://localhost:8080/health
```

## Building Custom Images

To build from source:

```bash
make docker
```

This creates a `rootgg/plik:dev` image with the current codebase.
